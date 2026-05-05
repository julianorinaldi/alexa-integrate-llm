package alexallm

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/joho/godotenv"
	_ "modernc.org/sqlite"
)

// AuthorizedSkill representa um registro no banco de dados
type AuthorizedSkill struct {
	ID          string
	SkillID     string
	SecretToken string
	OwnerName   string
}

var (
	db     *sql.DB
	dbOnce sync.Once
)

func getDB() *sql.DB {
	dbOnce.Do(func() {
		path := "/data/alexa.db"

		var err error
		db, err = sql.Open("sqlite", path)
		if err != nil {
			log.Fatalf("Erro ao abrir SQLite em %s: %v", path, err)
		}
		_, err = db.Exec(`CREATE TABLE IF NOT EXISTS authorized_skills (
			id          TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
			skill_id    TEXT NOT NULL,
			secret_token TEXT NOT NULL,
			owner_name  TEXT NOT NULL
		)`)
		if err != nil {
			log.Fatalf("Erro ao criar tabela authorized_skills: %v", err)
		}

		// Tabela de usuários do painel
		_, err = db.Exec(`CREATE TABLE IF NOT EXISTS dashboard_users (
			username TEXT PRIMARY KEY,
			password TEXT NOT NULL,
			must_change_password INTEGER DEFAULT 0
		)`)
		if err != nil {
			log.Fatalf("Erro ao criar tabela dashboard_users: %v", err)
		}

		// Cria usuário inicial admin:admin se a tabela estiver vazia
		var userCount int
		err = db.QueryRow("SELECT COUNT(*) FROM dashboard_users").Scan(&userCount)
		if err == nil && userCount == 0 {
			_, err = db.Exec("INSERT INTO dashboard_users (username, password, must_change_password) VALUES (?, ?, ?)", "admin", "admin", 1)
			if err != nil {
				log.Printf("⚠️ Aviso: Falha ao criar usuário inicial no SQLite: %v", err)
			} else {
				log.Printf("✅ Usuário padrão 'admin:admin' criado. Troca de senha obrigatória no primeiro acesso.")
			}
		}
	})
	return db
}

// isAuthorized verifica se a chamada da Alexa tem permissão (Env ou SQLite)
func isAuthorized(skillID, token string) bool {
	// 1. Verificação legada no .env
	envIDs := os.Getenv("ALEXA_SKILL_ID")
	envSecret := os.Getenv("ALEXA_SECRET_TOKEN")

	if token != "" && token == envSecret {
		for _, id := range strings.Split(envIDs, ",") {
			if strings.TrimSpace(id) == skillID {
				return true
			}
		}
	}

	// 2. Verificação no SQLite
	var count int
	err := getDB().QueryRow(
		`SELECT COUNT(*) FROM authorized_skills WHERE skill_id = ? AND secret_token = ?`,
		skillID, token,
	).Scan(&count)
	if err != nil {
		log.Printf("Erro consulta SQLite: %v", err)
		return false
	}
	return count > 0
}

func baseURL() string {
	base := os.Getenv("FUNCTION_BASE_URL")
	if base == "" {
		base = "http://localhost:8080"
	}
	return strings.TrimSuffix(base, "/")
}

func adminRedirect(w http.ResponseWriter, r *http.Request, page string) {
	http.Redirect(w, r, baseURL()+"/"+page, http.StatusSeeOther)
}

func handleAdminRouting(w http.ResponseWriter, r *http.Request) {
	godotenv.Load("../.env")
	godotenv.Load(".env")

	path := r.URL.Path

	if strings.HasSuffix(path, "/login") {
		handleLogin(w, r)
		return
	}

	if strings.HasSuffix(path, "/change-password") {
		handleChangePassword(w, r)
		return
	}

	if !isAuthenticated(r) {
		adminRedirect(w, r, "login")
		return
	}

	if strings.HasSuffix(path, "/delete") {
		handleDelete(w, r)
		return
	}

	if r.Method == http.MethodPost {
		handleAdd(w, r)
		return
	}

	renderDashboard(w)
}

func isAuthenticated(r *http.Request) bool {
	cookie, err := r.Cookie("admin_session")
	if err != nil {
		return false
	}
	return cookie.Value == "authorized"
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		renderLogin(w, "")
		return
	}

	user := r.FormValue("username")
	pass := r.FormValue("password")

	var dbUser string
	var mustChange int
	err := getDB().QueryRow(
		"SELECT username, must_change_password FROM dashboard_users WHERE username = ? AND password = ?",
		user, pass,
	).Scan(&dbUser, &mustChange)

	if err == nil && dbUser != "" {
		if mustChange == 1 {
			// Redireciona para troca de senha passando o user via cookie temporário
			http.SetCookie(w, &http.Cookie{
				Name:    "change_pwd_user",
				Value:   dbUser,
				Path:    "/",
				Expires: time.Now().Add(5 * time.Minute),
			})
			adminRedirect(w, r, "change-password")
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:    "admin_session",
			Value:   "authorized",
			Path:    "/",
			Expires: time.Now().Add(24 * time.Hour),
		})
		adminRedirect(w, r, "admin")
		return
	}

	log.Printf("Tentativa de login falhou para o usuário: '%s'", user)
	renderLogin(w, "Credenciais inválidas.")
}

func handleChangePassword(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("change_pwd_user")
	if err != nil {
		adminRedirect(w, r, "login")
		return
	}
	username := cookie.Value

	if r.Method == http.MethodGet {
		renderChangePassword(w, username, "")
		return
	}

	newPass := r.FormValue("new_password")
	confirmPass := r.FormValue("confirm_password")

	if newPass == "" || newPass == "admin" {
		renderChangePassword(w, username, "Escolha uma senha diferente de 'admin'.")
		return
	}

	if newPass != confirmPass {
		renderChangePassword(w, username, "As senhas não coincidem.")
		return
	}

	_, err = getDB().Exec(
		"UPDATE dashboard_users SET password = ?, must_change_password = 0 WHERE username = ?",
		newPass, username,
	)
	if err != nil {
		renderChangePassword(w, username, "Erro ao salvar nova senha.")
		return
	}

	// Limpa cookie de troca e loga
	http.SetCookie(w, &http.Cookie{Name: "change_pwd_user", MaxAge: -1, Path: "/"})
	http.SetCookie(w, &http.Cookie{
		Name:    "admin_session",
		Value:   "authorized",
		Path:    "/",
		Expires: time.Now().Add(24 * time.Hour),
	})
	adminRedirect(w, r, "admin")
}

func renderChangePassword(w http.ResponseWriter, username, errorMsg string) {
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintf(w, `
	<!DOCTYPE html>
	<html lang="pt-br">
	<head>
		<meta charset="UTF-8">
		<meta name="viewport" content="width=device-width, initial-scale=1.0">
		<title>Alterar Senha - Alexa Admin</title>
		<style>
			:root { --bg: #0f172a; --accent: #38bdf8; }
			body { background: var(--bg); color: white; font-family: sans-serif; display: flex; align-items: center; justify-content: center; height: 100vh; margin: 0; }
			.card { background: rgba(30, 41, 59, 0.7); backdrop-filter: blur(10px); padding: 2rem; border-radius: 1rem; border: 1px solid rgba(255,255,255,0.1); width: 320px; }
			input { width: 100%%; padding: 0.75rem; margin-bottom: 1rem; background: rgba(0,0,0,0.2); border: 1px solid rgba(255,255,255,0.1); border-radius: 0.5rem; color: white; box-sizing: border-box; }
			button { width: 100%%; padding: 0.75rem; background: var(--accent); border: none; border-radius: 0.5rem; font-weight: bold; cursor: pointer; }
			.error { color: #f87171; font-size: 0.8rem; margin-bottom: 1rem; text-align: center; }
		</style>
	</head>
	<body>
		<div class="card">
			<h2 style="text-align:center">Nova Senha</h2>
			<p style="font-size:0.9rem; color:#94a3b8; text-align:center">Olá <b>%s</b>, por segurança você deve alterar sua senha inicial.</p>
			%s
			<form method="POST">
				<input type="password" name="new_password" placeholder="Nova Senha" required>
				<input type="password" name="confirm_password" placeholder="Confirme a Nova Senha" required>
				<button>Salvar e Entrar</button>
			</form>
		</div>
	</body>
	</html>`, username, formatError(errorMsg))
}

func renderLogin(w http.ResponseWriter, errorMsg string) {
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintf(w, `
	<!DOCTYPE html>
	<html lang="pt-br">
	<head>
		<meta charset="UTF-8">
		<meta name="viewport" content="width=device-width, initial-scale=1.0">
		<title>Login - Alexa Admin</title>
		<style>
			:root { --bg: #0f172a; --accent: #38bdf8; }
			body { background: var(--bg); color: white; font-family: sans-serif; display: flex; align-items: center; justify-content: center; height: 100vh; margin: 0; }
			.card { background: rgba(30, 41, 59, 0.7); backdrop-filter: blur(10px); padding: 2rem; border-radius: 1rem; border: 1px solid rgba(255,255,255,0.1); width: 300px; }
			input { width: 100%%; padding: 0.75rem; margin-bottom: 1rem; background: rgba(0,0,0,0.2); border: 1px solid rgba(255,255,255,0.1); border-radius: 0.5rem; color: white; box-sizing: border-box; }
			button { width: 100%%; padding: 0.75rem; background: var(--accent); border: none; border-radius: 0.5rem; font-weight: bold; cursor: pointer; }
			.error { color: #f87171; font-size: 0.8rem; margin-bottom: 1rem; text-align: center; }
		</style>
	</head>
	<body>
		<div class="card">
			<h2 style="text-align:center">Painel Alexa</h2>
			%s
			<form method="POST" action="%s/login"><input type="text" name="username" placeholder="Usuário"><input type="password" name="password" placeholder="Senha"><button>Entrar</button></form>
		</div>
	</body>
	</html>`, formatError(errorMsg), baseURL())
}

func formatError(msg string) string {
	if msg == "" {
		return ""
	}
	return fmt.Sprintf(`<div class="error">%s</div>`, msg)
}

func renderDashboard(w http.ResponseWriter) {
	skills, _ := getSkillsFromDB()

	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintf(w, `
	<!DOCTYPE html>
	<html>
	<head>
		<meta charset="UTF-8">
		<title>Dashboard Alexa</title>
		<style>
			body { background: #0f172a; color: white; font-family: sans-serif; padding: 2rem; }
			.container { max-width: 900px; margin: 0 auto; }
			.card { background: rgba(30, 41, 59, 0.7); padding: 1.5rem; border-radius: 1rem; border: 1px solid rgba(255,255,255,0.1); margin-bottom: 2rem; }
			table { width: 100%%; border-collapse: collapse; }
			th, td { text-align: left; padding: 1rem; border-bottom: 1px solid rgba(255,255,255,0.05); }
			input { background: rgba(0,0,0,0.2); border: 1px solid rgba(255,255,255,0.1); padding: 0.5rem; border-radius: 0.3rem; color: white; margin-right: 0.5rem; }
			button { background: #38bdf8; border: none; padding: 0.5rem 1rem; border-radius: 0.3rem; cursor: pointer; font-weight: bold; }
		</style>
	</head>
	<body>
		<div class="container">
			<header>
				<h1>Painel de Acessos</h1>
				<a href="./" style="color: grey; text-decoration: none;">Sair</a>
			</header>
			<div class="card">
				<h3>Novo Dispositivo/Pessoa</h3>
				<form method="POST" action="%s/admin">
					<input type="text" name="owner" placeholder="Nome" required>
					<input type="text" name="skill_id" placeholder="Skill ID" required>
					<input type="text" name="token" placeholder="Token" required>
					<button>Adicionar</button>
				</form>
			</div>
			<div class="card">
				<table>
					<thead><tr><th>Dono</th><th>Skill ID</th><th>Token</th><th>Ação</th></tr></thead>
					<tbody>%s</tbody>
				</table>
			</div>
		</div>
	</body>
	</html>`, baseURL(), renderTableRows(skills, baseURL()))
}

func renderTableRows(skills []AuthorizedSkill, base string) string {
	var res string
	for _, s := range skills {
		res += fmt.Sprintf(`<tr><td>%s</td><td>%s</td><td>%s</td><td><a href="%s/admin/delete?id=%s" style="color:#f87171">Remover</a></td></tr>`,
			s.OwnerName, s.SkillID, s.SecretToken, base, s.ID)
	}
	return res
}

func getSkillsFromDB() ([]AuthorizedSkill, error) {
	rows, err := getDB().Query(`SELECT id, skill_id, secret_token, owner_name FROM authorized_skills`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var skills []AuthorizedSkill
	for rows.Next() {
		var s AuthorizedSkill
		if err := rows.Scan(&s.ID, &s.SkillID, &s.SecretToken, &s.OwnerName); err != nil {
			return nil, err
		}
		skills = append(skills, s)
	}
	return skills, rows.Err()
}

func handleAdd(w http.ResponseWriter, r *http.Request) {
	_, err := getDB().Exec(
		`INSERT INTO authorized_skills (id, skill_id, secret_token, owner_name)
		 VALUES (lower(hex(randomblob(16))), ?, ?, ?)`,
		r.FormValue("skill_id"), r.FormValue("token"), r.FormValue("owner"),
	)
	if err != nil {
		log.Printf("Erro ao inserir skill: %v", err)
	}
	adminRedirect(w, r, "admin")
}

func handleDelete(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	_, err := getDB().Exec(`DELETE FROM authorized_skills WHERE id = ?`, id)
	if err != nil {
		log.Printf("Erro ao deletar skill: %v", err)
	}
	adminRedirect(w, r, "admin")
}
