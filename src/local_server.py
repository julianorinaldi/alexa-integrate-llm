import os
from flask import Flask, request, jsonify
from ask_sdk_webservice_support.webservice_handler import WebserviceSkillHandler
from src.main import sb  # Nossa SkillBuilder configurada

app = Flask(__name__)

# Adaptador que traduz requisições HTTP para a Skill
handler = WebserviceSkillHandler(skill=sb.create())

@app.route("/", methods=["POST"])
def invoke_skill():
    """
    Este endpoint recebe o POST da Alexa (ou simulações via HTTP).
    """
    return handler.verify_request_and_dispatch(
        http_request_headers=request.headers,
        http_request_body=request.data.decode("utf-8")
    )

if __name__ == "__main__":
    # Rodar servidor na porta 5000 (Exposta pelo Docker)
    port = int(os.environ.get("PORT", 5000))
    print(f"📡 Servidor de Debug Local rodando na porta {port}...")
    app.run(host="0.0.0.0", port=port, debug=True)
