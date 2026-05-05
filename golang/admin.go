package alexallm

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	_ "modernc.org/sqlite"
)

// AuthorizedSkill representa um registro no banco de dados
type AuthorizedSkill struct {
	ID          string
	SkillID     string
	SecretToken string
	OwnerName   string
	Description string
}

// LLMConfig representa a configuração do modelo de linguagem
type LLMConfig struct {
	Name        string
	APIKey      string
	ModelName   string
	Description string
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
			owner_name  TEXT NOT NULL,
			description TEXT DEFAULT ''
		)`)
		if err != nil {
			log.Fatalf("Erro ao criar tabela authorized_skills: %v", err)
		}
		// Atualiza tabela antiga caso já exista
		db.Exec(`ALTER TABLE authorized_skills ADD COLUMN description TEXT DEFAULT ''`)

		// Tabela de Configuração do LLM
		_, err = db.Exec(`CREATE TABLE IF NOT EXISTS llm_config (
			id INTEGER PRIMARY KEY CHECK (id = 1),
			name TEXT,
			api_key TEXT,
			model_name TEXT,
			description TEXT
		)`)
		if err == nil {
			db.Exec(`INSERT OR IGNORE INTO llm_config (id, name, api_key, model_name, description) VALUES (1, 'OpenRouter Padrão', '', 'google/gemini-2.5-flash-lite', 'Configuração inicial da IA')`)
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

// InitDB força a inicialização do banco de dados imediatamente (útil no boot do servidor local)
func InitDB() {
	getDB()
}

// GetLLMConfig retorna as configs do LLM salvas no banco
func GetLLMConfig() (string, string) {
	var apiKey, modelName string
	err := getDB().QueryRow(`SELECT api_key, model_name FROM llm_config WHERE id = 1`).Scan(&apiKey, &modelName)
	if err != nil {
		log.Printf("Erro ao buscar LLM config: %v", err)
		return "", "openai/gpt-3.5-turbo"
	}
	return apiKey, modelName
}

// isAuthorized verifica se a chamada da Alexa tem permissão (Somente SQLite)
func isAuthorized(skillID, token string) bool {
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

func adminRedirect(w http.ResponseWriter, r *http.Request, page string) {
	http.Redirect(w, r, page, http.StatusSeeOther)
}

func handleAdminRouting(w http.ResponseWriter, r *http.Request) {
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

	if strings.HasSuffix(path, "/llm-config") {
		handleSaveLLM(w, r)
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
			<form method="POST" action="login"><input type="text" name="username" placeholder="Usuário"><input type="password" name="password" placeholder="Senha"><button>Entrar</button></form>
		</div>
	</body>
	</html>`, formatError(errorMsg))
}

func formatError(msg string) string {
	if msg == "" {
		return ""
	}
	return fmt.Sprintf(`<div class="error">%s</div>`, msg)
}

func renderDashboard(w http.ResponseWriter) {
	skills, _ := getSkillsFromDB()
	var llm LLMConfig
	getDB().QueryRow(`SELECT name, api_key, model_name, description FROM llm_config WHERE id = 1`).Scan(&llm.Name, &llm.APIKey, &llm.ModelName, &llm.Description)

	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintf(w, `
	<!DOCTYPE html>
	<html lang="pt-br">
	<head>
		<meta charset="UTF-8">
		<meta name="viewport" content="width=device-width, initial-scale=1.0">
		<title>Dashboard Alexa</title>
		<style>
			:root { --bg: #0f172a; --card-bg: rgba(30, 41, 59, 0.7); --border: rgba(255,255,255,0.1); --accent: #38bdf8; --text: #f1f5f9; --text-muted: #94a3b8; }
			body { background: var(--bg); color: var(--text); font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif; padding: 2rem; margin: 0; }
			.container { max-width: 950px; margin: 0 auto; }
			
			header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 2rem; padding-bottom: 1rem; border-bottom: 1px solid var(--border); }
			header h1 { margin: 0; font-size: 1.8rem; font-weight: 600; }
			.btn-logout { color: #f87171; text-decoration: none; font-weight: bold; padding: 0.5rem 1rem; border: 1px solid rgba(248,113,113,0.3); border-radius: 0.5rem; transition: all 0.2s; }
			.btn-logout:hover { background: rgba(248,113,113,0.1); }
			
			.card { background: var(--card-bg); backdrop-filter: blur(10px); padding: 1.5rem 2rem; border-radius: 1rem; border: 1px solid var(--border); margin-bottom: 2rem; box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1); }
			.card h3 { margin-top: 0; color: var(--accent); font-size: 1.2rem; display: flex; align-items: center; gap: 0.5rem; }
			
			label { display: block; font-size: 0.85rem; color: var(--text-muted); margin-bottom: 0.3rem; font-weight: 500; }
			input { width: 100%%; background: rgba(0,0,0,0.25); border: 1px solid var(--border); padding: 0.75rem; border-radius: 0.5rem; color: white; margin-bottom: 1.2rem; box-sizing: border-box; transition: border-color 0.2s; }
			input:focus { outline: none; border-color: var(--accent); }
			
			button { background: var(--accent); color: #0f172a; border: none; padding: 0.75rem 1.5rem; border-radius: 0.5rem; cursor: pointer; font-weight: bold; font-size: 0.95rem; transition: transform 0.1s, opacity 0.2s; }
			button:hover { opacity: 0.9; }
			button:active { transform: scale(0.98); }
			
			.form-row { display: grid; grid-template-columns: repeat(auto-fit, minmax(180px, 1fr)); gap: 1rem; align-items: start; }
			.form-row .input-group { margin-bottom: 0; }
			.form-row input { margin-bottom: 0; }
			.form-row button { height: 100%%; min-height: 42px; margin-top: 1.2rem; }
			
			table { width: 100%%; border-collapse: collapse; margin-top: 1rem; }
			th { text-align: left; padding: 1rem; border-bottom: 2px solid var(--border); color: var(--text-muted); font-size: 0.85rem; text-transform: uppercase; letter-spacing: 0.05em; }
			td { padding: 1rem; border-bottom: 1px solid rgba(255,255,255,0.05); word-break: break-all; }
			tr:last-child td { border-bottom: none; }
			tr:hover td { background: rgba(255,255,255,0.02); }
			
			.action-link { color: #f87171; text-decoration: none; font-weight: 600; font-size: 0.9rem; }
			.action-link:hover { text-decoration: underline; }
		</style>
	</head>
	<body>
		<div class="container">
			<header>
				<h1>Painel de Acessos</h1>
				<a href="./" class="btn-logout">Sair do Painel</a>
			</header>

			<!-- Seção de Configuração LLM -->
			<div class="card">
				<h3>🤖 Configuração do Modelo de IA</h3>
				<form method="POST" action="admin/llm-config">
					<label>Nome do Serviço</label>
					<input type="text" name="name" value="%s" placeholder="Ex: OpenRouter Principal" required>
					
					<label>Chave de API (OPENROUTER_API_KEY)</label>
					<input type="password" name="api_key" value="%s" placeholder="sk-or-v1-..." required>
					
					<label>Modelo (MODEL_NAME)</label>
					<input type="text" name="model_name" value="%s" placeholder="Ex: google/gemini-2.5-flash-lite" required>
					
					<label>Descrição Opcional</label>
					<input type="text" name="description" value="%s" placeholder="Detalhes do uso">
					
					<div style="text-align: right;">
						<button>Salvar Configurações LLM</button>
					</div>
				</form>
			</div>

			<!-- Seção de Skills -->
			<div class="card">
				<h3>🗣️ Adicionar Novo Dispositivo / Skill</h3>
				<form method="POST" action="admin">
					<div class="form-row">
						<div class="input-group">
							<label>Dono (Identificação)</label>
							<input type="text" name="owner" placeholder="Nome" required>
						</div>
						<div class="input-group">
							<label>Skill ID</label>
							<input type="text" name="skill_id" placeholder="amzn1.ask.skill..." required>
						</div>
						<div class="input-group">
							<label>Token de Segurança</label>
							<input type="text" name="token" placeholder="Senha única" required>
						</div>
						<div class="input-group">
							<label>Descrição</label>
							<input type="text" name="description" placeholder="Onde está rodando?">
						</div>
						<button>Adicionar</button>
					</div>
				</form>
			</div>
			
			<div class="card" style="padding: 1rem 0;">
				<h3 style="padding: 0 2rem;">📋 Dispositivos Autorizados</h3>
				<div style="overflow-x: auto;">
					<table>
						<thead><tr><th>Dono</th><th>Skill ID</th><th>Token</th><th>Descrição</th><th style="text-align:center">Ação</th></tr></thead>
						<tbody>%s</tbody>
					</table>
				</div>
			</div>
		</div>
	</body>
	</html>`, llm.Name, llm.APIKey, llm.ModelName, llm.Description, renderTableRows(skills))
}

func renderTableRows(skills []AuthorizedSkill) string {
	var res string
	for _, s := range skills {
		res += fmt.Sprintf(`<tr><td>%s</td><td>%s</td><td>%s</td><td>%s</td><td style="text-align:center"><a href="admin/delete?id=%s" class="action-link">Remover</a></td></tr>`,
			s.OwnerName, s.SkillID, s.SecretToken, s.Description, s.ID)
	}
	return res
}

func getSkillsFromDB() ([]AuthorizedSkill, error) {
	rows, err := getDB().Query(`SELECT id, skill_id, secret_token, owner_name, description FROM authorized_skills`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var skills []AuthorizedSkill
	for rows.Next() {
		var s AuthorizedSkill
		if err := rows.Scan(&s.ID, &s.SkillID, &s.SecretToken, &s.OwnerName, &s.Description); err != nil {
			return nil, err
		}
		skills = append(skills, s)
	}
	return skills, rows.Err()
}

func handleAdd(w http.ResponseWriter, r *http.Request) {
	_, err := getDB().Exec(
		`INSERT INTO authorized_skills (id, skill_id, secret_token, owner_name, description)
		 VALUES (lower(hex(randomblob(16))), ?, ?, ?, ?)`,
		r.FormValue("skill_id"), r.FormValue("token"), r.FormValue("owner"), r.FormValue("description"),
	)
	if err != nil {
		log.Printf("Erro ao inserir skill: %v", err)
	}
	adminRedirect(w, r, "admin")
}

func handleSaveLLM(w http.ResponseWriter, r *http.Request) {
	_, err := getDB().Exec(
		`UPDATE llm_config SET name = ?, api_key = ?, model_name = ?, description = ? WHERE id = 1`,
		r.FormValue("name"), r.FormValue("api_key"), r.FormValue("model_name"), r.FormValue("description"),
	)
	if err != nil {
		log.Printf("Erro ao atualizar llm config: %v", err)
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
