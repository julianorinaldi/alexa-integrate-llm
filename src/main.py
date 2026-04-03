from ask_sdk_core.skill_builder import SkillBuilder
from src.alexa.handlers import (
    LaunchRequestHandler, 
    AskIntentHandler, 
    CancelOrStopIntentHandler
)

# Registramos os handlers na skill
sb = SkillBuilder()

# Adicionados por ordem de prioridade
sb.add_request_handler(LaunchRequestHandler())
sb.add_request_handler(AskIntentHandler())
sb.add_request_handler(CancelOrStopIntentHandler())

# Exportamos para o AWS Lambda invocar
handler = sb.lambda_handler()
