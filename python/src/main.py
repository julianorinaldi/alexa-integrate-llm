from ask_sdk_core.skill_builder import SkillBuilder
from src.alexa.handlers import (
    LaunchRequestHandler, 
    AskIntentHandler, 
    CancelOrStopIntentHandler,
    SessionEndedRequestHandler,
    CatchAllRequestHandler,
    SkillIdVerifierInterceptor
)

import os

# Registramos os handlers na skill
sb = SkillBuilder()

# Adicionados por ordem de prioridade
sb.add_request_handler(LaunchRequestHandler())
sb.add_request_handler(AskIntentHandler())
sb.add_request_handler(CancelOrStopIntentHandler())
sb.add_request_handler(SessionEndedRequestHandler())
sb.add_request_handler(CatchAllRequestHandler())

# Adicionamos o interceptor para validar múltiplos Skill IDs
sb.add_global_request_interceptor(SkillIdVerifierInterceptor())

# Exportamos para o AWS Lambda invocar
handler = sb.lambda_handler()
