import pytest
from ask_sdk_core.handler_input import HandlerInput
from ask_sdk_model import RequestEnvelope, Response, LaunchRequest
from src.alexa.handlers import LaunchRequestHandler
from unittest.mock import MagicMock

def test_launch_request_handler_can_handle():
    # Cria mock de HandlerInput
    request_envelope = RequestEnvelope(
        request=LaunchRequest(
            locale="pt-BR",
            request_id="amzn1.echo-api.request.000",
            timestamp="2024-05-01T00:00:00Z"
        )
    )
    handler_input = HandlerInput(request_envelope=request_envelope)
    
    handler = LaunchRequestHandler()
    assert handler.can_handle(handler_input) == True

def test_launch_request_handler_handle_returns_speech():
    # Cria mock de HandlerInput con atributos
    request_envelope = RequestEnvelope(
        request=LaunchRequest(
            locale="pt-BR"
        ),
        context=MagicMock()
    )
    
    # Mock do AttributesManager que a Alexa SDK usa
    attr_manager = MagicMock()
    attr_manager.session_attributes = {}
    
    handler_input = HandlerInput(
        request_envelope=request_envelope,
        attributes_manager=attr_manager
    )
    
    handler = LaunchRequestHandler()
    response = handler.handle(handler_input)
    
    # Verifica se a mensagem de saída está correta (conforme no handler)
    assert response.output_speech.ssml.startswith("<speak>Ol")
