package alexallm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// AuthorizedSkill representa um registro no banco de dados
type AuthorizedSkill struct {
	ID          string `json:"id,omitempty"`
	SkillID     string `json:"skill_id"`
	SecretToken string `json:"secret_token"`
	OwnerName   string `json:"owner_name"`
}

// isAuthorized verifica se a chamada da Alexa tem permissão (Env ou Supabase)
func isAuthorized(skillID, token string) bool {
	// 1. Verificação legada no .env
	envIDs := os.Getenv("ALEXA_SKILL_ID")
	envSecret := os.Getenv("ALEXA_SECRET_TOKEN")
	
	// Se o token for passado e bater com o segredo global, verificamos os IDs do env
	if token != "" && token == envSecret {
		for _, id := range strings.Split(envIDs, ",") {
			if strings.TrimSpace(id) == skillID {
				return true
			}
		}
	}

	// 2. Verificação no Supabase
	return checkSupabaseAuth(skillID, token)
}

func checkSupabaseAuth(skillID, token string) bool {
	supabaseURL := os.Getenv("SUPABASE_URL")
	supabaseKey := os.Getenv("SUPABASE_SERVICE_ROLE_KEY")
	if supabaseURL == "" || supabaseKey == "" || strings.Contains(supabaseURL, "sua-url") {
		return false
	}

	url := fmt.Sprintf("%s/rest/v1/authorized_skills?skill_id=eq.%s&secret_token=eq.%s&select=count", supabaseURL, skillID, token)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("apikey", supabaseKey)
	req.Header.Set("Authorization", "Bearer "+supabaseKey)
	req.Header.Set("Range-Unit", "items")
	req.Header.Set("Range", "0-0")

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Erro consulta Supabase: %v", err)
		return false
	}
	defer resp.Body.Close()

	// Se encontrar 1 registro, está autorizado
	return resp.StatusCode == http.StatusOK && resp.Header.Get("Content-Range") != "" && !strings.HasSuffix(resp.Header.Get("Content-Range"), "/0")
}

// baseURL retorna a URL base do painel (ex: https://....cloudfunctions.net/alexa-llm-go)
func baseURL() string {
	base := os.Getenv("FUNCTION_BASE_URL")
	if base == "" {
		base = "http://localhost:8080" // fallback local
	}
	return strings.TrimSuffix(base, "/")
}

func adminRedirect(w http.ResponseWriter, r *http.Request, page string) {
	http.Redirect(w, r, baseURL()+"/"+page, http.StatusSeeOther)
}


func handleAdminRouting(w http.ResponseWriter, r *http.Request) {
	// Tentar carregar .env se existir (útil para desenvolvimento local e GCF se o arquivo for enviado)
	godotenv.Load("../.env")
	godotenv.Load(".env")

	path := r.URL.Path
	
	if strings.HasSuffix(path, "/login") {
		handleLogin(w, r)
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

	dashUser := os.Getenv("DASHBOARD_USER")
	dashPass := os.Getenv("DASHBOARD_PASS")

	if user != "" && user == dashUser && pass == dashPass {
		http.SetCookie(w, &http.Cookie{
			Name:    "admin_session",
			Value:   "authorized",
			Path:    "/",
			Expires: time.Now().Add(24 * time.Hour),
		})
		adminRedirect(w, r, "admin")
		return
	}

	log.Printf("Tentativa de login falhou. Env User: '%s', Pass configured: %v", dashUser, dashPass != "")
	renderLogin(w, "Credenciais inválidas ou variáveis não carregadas.")
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
	if msg == "" { return "" }
	return fmt.Sprintf(`<div class="error">%s</div>`, msg)
}

func renderDashboard(w http.ResponseWriter) {
	skills, _ := getSkillsFromSupabase()
	
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
		res += fmt.Sprintf(`<tr><td>%s</td><td>%s</td><td>%s</td><td><a href="%s/admin/delete?id=%s" style="color:#f87171">Remover</a></td></tr>`, s.OwnerName, s.SkillID, s.SecretToken, base, s.ID)
	}
	return res
}

func getSkillsFromSupabase() ([]AuthorizedSkill, error) {
	supabaseURL := os.Getenv("SUPABASE_URL")
	supabaseKey := os.Getenv("SUPABASE_SERVICE_ROLE_KEY")
	if supabaseURL == "" { return nil, nil }
	req, _ := http.NewRequest("GET", supabaseURL+"/rest/v1/authorized_skills?select=*", nil)
	req.Header.Set("apikey", supabaseKey)
	req.Header.Set("Authorization", "Bearer "+supabaseKey)
	client := &http.Client{Timeout: 5 * time.Second}
	resp, _ := client.Do(req)
	if resp == nil { return nil, nil }
	defer resp.Body.Close()
	var skills []AuthorizedSkill
	json.NewDecoder(resp.Body).Decode(&skills)
	return skills, nil
}

func handleAdd(w http.ResponseWriter, r *http.Request) {
	skill := AuthorizedSkill{
		OwnerName:   r.FormValue("owner"),
		SkillID:     r.FormValue("skill_id"),
		SecretToken: r.FormValue("token"),
	}
	body, _ := json.Marshal(skill)
	supabaseURL := os.Getenv("SUPABASE_URL")
	supabaseKey := os.Getenv("SUPABASE_SERVICE_ROLE_KEY")
	req, _ := http.NewRequest("POST", supabaseURL+"/rest/v1/authorized_skills", bytes.NewBuffer(body))
	req.Header.Set("apikey", supabaseKey)
	req.Header.Set("Authorization", "Bearer "+supabaseKey)
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{Timeout: 5 * time.Second}
	client.Do(req)
	adminRedirect(w, r, "admin")
}

func handleDelete(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	supabaseURL := os.Getenv("SUPABASE_URL")
	supabaseKey := os.Getenv("SUPABASE_SERVICE_ROLE_KEY")
	req, _ := http.NewRequest("DELETE", supabaseURL+"/rest/v1/authorized_skills?id=eq."+id, nil)
	req.Header.Set("apikey", supabaseKey)
	req.Header.Set("Authorization", "Bearer "+supabaseKey)
	client := &http.Client{Timeout: 5 * time.Second}
	client.Do(req)
	adminRedirect(w, r, "admin")
}
