import json
import os
from ask_sdk_core.dispatch_components import AbstractRequestHandler, AbstractRequestInterceptor
from ask_sdk_core.utils import is_request_type, is_intent_name
from ask_sdk_core.handler_input import HandlerInput
from ask_sdk_model import Response
from ask_sdk_core.exceptions import AlexaSdkException
from ask_sdk_model.interfaces.alexa.presentation.apl import RenderDocumentDirective
from src.llm.client import llm_client

class LaunchRequestHandler(AbstractRequestHandler):
    """Handler para quando a skill é aberta ('Alexa, abra meu assistente')."""
    def can_handle(self, handler_input: HandlerInput) -> bool:
        return is_request_type("LaunchRequest")(handler_input)

    def handle(self, handler_input: HandlerInput) -> Response:
        speak_output = "Olá! Sou seu assistente inteligente. O que deseja me perguntar hoje?"
        
        # Iniciando histórico na sessão (Memória que vovê pediu)
        session_attr = handler_input.attributes_manager.session_attributes
        session_attr["chat_history"] = []
        handler_input.attributes_manager.session_attributes = session_attr

        # Adicionando visual APL se for um Echo Show
        if handler_input.request_envelope.context.system.device.supported_interfaces.alexa_presentation_apl:
            apl_document = self._load_apl_template("welcome_template.json")
            handler_input.response_builder.add_directive(
                RenderDocumentDirective(
                    document=apl_document,
                    datasources={
                        "welcomeData": {
                            "header": "Bem-vindo!",
                            "text": "Estou pronto para ouvir. Faça sua pergunta."
                        }
                    }
                )
            )

        return (
            handler_input.response_builder
                .speak(speak_output)
                .ask(speak_output) # Mantém a skill aberta esperando resposta
                .response
        )

    def _load_apl_template(self, filename: str):
        path = os.path.join(os.path.dirname(__file__), "templates", filename)
        with open(path, "r") as f:
            return json.load(f)

class AskIntentHandler(AbstractRequestHandler):
    """Handler principal que recebe a pergunta do usuário e envia para a LLM."""
    def can_handle(self, handler_input: HandlerInput) -> bool:
        return is_intent_name("AskIntent")(handler_input)

    def handle(self, handler_input: HandlerInput) -> Response:
        # Pegar a pergunta do slot 'query' (precisa configurar no console da Alexa)
        slots = handler_input.request_envelope.request.intent.slots
        user_query = slots["query"].value if "query" in slots else "O que você pode fazer?"

        # Recuperar histórico da sessão (Contexto)
        session_attr = handler_input.attributes_manager.session_attributes
        history = session_attr.get("chat_history", [])

        # Chamar a LLM
        answer = llm_client.ask(user_query, history=history)

        # Atualizar histórico com a nova interação (Contexto de Memória)
        history.append({"role": "user", "content": user_query})
        history.append({"role": "assistant", "content": answer})
        session_attr["chat_history"] = history
        handler_input.attributes_manager.session_attributes = session_attr

        # Preparar visual APL
        if handler_input.request_envelope.context.system.device.supported_interfaces.alexa_presentation_apl:
            apl_document = self._load_apl_template("response_template.json")
            handler_input.response_builder.add_directive(
                RenderDocumentDirective(
                    document=apl_document,
                    datasources={
                        "chatData": {
                            "question": user_query,
                            "answer": answer
                        }
                    }
                )
            )

        return (
            handler_input.response_builder
                .speak(answer)
                .ask("Deseja me perguntar mais alguma coisa?") 
                .response
        )

    def _load_apl_template(self, filename: str):
        path = os.path.join(os.path.dirname(__file__), "templates", filename)
        with open(path, "r") as f:
            return json.load(f)

class CancelOrStopIntentHandler(AbstractRequestHandler):
    def can_handle(self, handler_input: HandlerInput) -> bool:
        return (is_intent_name("AMAZON.CancelIntent")(handler_input) or
                is_intent_name("AMAZON.StopIntent")(handler_input))

    def handle(self, handler_input: HandlerInput) -> Response:
        return (
            handler_input.response_builder
                .speak("Até logo! Estarei aqui se precisar.")
                .set_should_end_session(True)
                .response
        )

class SessionEndedRequestHandler(AbstractRequestHandler):
    def can_handle(self, handler_input: HandlerInput) -> bool:
        return is_request_type("SessionEndedRequest")(handler_input)

    def handle(self, handler_input: HandlerInput) -> Response:
        # Apenas aceita sem erro
        return handler_input.response_builder.response

class CatchAllRequestHandler(AbstractRequestHandler):
    """Fallback genérico para intenções ou eventos que não configuramos."""
    def can_handle(self, handler_input: HandlerInput) -> bool:
        return True

    def handle(self, handler_input: HandlerInput) -> Response:
        req = handler_input.request_envelope.request
        if hasattr(req, "intent"):
            intent_name = req.intent.name
            speak_output = f"A intenção {intent_name} foi chamada, mas não tenho um handler para ela."
        else:
            speak_output = f"Recebi um evento do tipo {req.type}, que não sei como processar."
            
        return (
            handler_input.response_builder
                .speak(speak_output)
                .ask(speak_output)
                .response
        )

class SkillIdVerifierInterceptor(AbstractRequestInterceptor):
    """Verifica se o ID da Skill que está chamando está na lista permitida."""
    def process(self, handler_input):
        # type: (HandlerInput) -> None
        expected_ids_str = os.getenv("ALEXA_SKILL_ID")
        if not expected_ids_str:
            return

        allowed_ids = [s.strip() for s in expected_ids_str.split(",")]
        current_id = handler_input.request_envelope.context.system.application.application_id
        
        if current_id not in allowed_ids:
            print(f"Acesso negado: Skill ID {current_id} não autorizado.")
            raise AlexaSdkException("Skill ID não autorizado.")
