from ask_sdk_core.skill_builder import SkillBuilder
from src.alexa.handlers import (
    LaunchRequestHandler, 
    AskIntentHandler, 
    CancelOrStopIntentHandler,
    SessionEndedRequestHandler,
    CatchAllRequestHandler
)

import os

# Registramos os handlers na skill
sb = SkillBuilder()
expected_skill_id = os.getenv("ALEXA_SKILL_ID")
if expected_skill_id:
    sb.skill_id = expected_skill_id

# Adicionados por ordem de prioridade
sb.add_request_handler(LaunchRequestHandler())
sb.add_request_handler(AskIntentHandler())
sb.add_request_handler(CancelOrStopIntentHandler())
sb.add_request_handler(SessionEndedRequestHandler())
sb.add_request_handler(CatchAllRequestHandler())

# Exportamos para o AWS Lambda invocar
handler = sb.lambda_handler()
