# Generated by the gRPC Python protocol compiler plugin. DO NOT EDIT!
"""Client and server classes corresponding to protobuf-defined services."""
import grpc

import medical_service_pb2 as medical__service__pb2


class MedicalChatServiceStub(object):
    """Main Chat Service
    """

    def __init__(self, channel):
        """Constructor.

        Args:
            channel: A grpc.Channel.
        """
        self.ChatStream = channel.stream_stream(
                '/backend.MedicalChatService/ChatStream',
                request_serializer=medical__service__pb2.ChatRequest.SerializeToString,
                response_deserializer=medical__service__pb2.ChatResponse.FromString,
                )


class MedicalChatServiceServicer(object):
    """Main Chat Service
    """

    def ChatStream(self, request_iterator, context):
        """Bidirectional streaming for real-time chat
        """
        context.set_code(grpc.StatusCode.UNIMPLEMENTED)
        context.set_details('Method not implemented!')
        raise NotImplementedError('Method not implemented!')


def add_MedicalChatServiceServicer_to_server(servicer, server):
    rpc_method_handlers = {
            'ChatStream': grpc.stream_stream_rpc_method_handler(
                    servicer.ChatStream,
                    request_deserializer=medical__service__pb2.ChatRequest.FromString,
                    response_serializer=medical__service__pb2.ChatResponse.SerializeToString,
            ),
    }
    generic_handler = grpc.method_handlers_generic_handler(
            'backend.MedicalChatService', rpc_method_handlers)
    server.add_generic_rpc_handlers((generic_handler,))


 # This class is part of an EXPERIMENTAL API.
class MedicalChatService(object):
    """Main Chat Service
    """

    @staticmethod
    def ChatStream(request_iterator,
            target,
            options=(),
            channel_credentials=None,
            call_credentials=None,
            insecure=False,
            compression=None,
            wait_for_ready=None,
            timeout=None,
            metadata=None):
        return grpc.experimental.stream_stream(request_iterator, target, '/backend.MedicalChatService/ChatStream',
            medical__service__pb2.ChatRequest.SerializeToString,
            medical__service__pb2.ChatResponse.FromString,
            options, channel_credentials,
            insecure, call_credentials, compression, wait_for_ready, timeout, metadata)


class MedicalQAServiceStub(object):
    """Main QA service definition
    """

    def __init__(self, channel):
        """Constructor.

        Args:
            channel: A grpc.Channel.
        """
        self.GenerateDraftAnswer = channel.unary_unary(
                '/backend.MedicalQAService/GenerateDraftAnswer',
                request_serializer=medical__service__pb2.QuestionRequest.SerializeToString,
                response_deserializer=medical__service__pb2.QuestionResponse.FromString,
                )


class MedicalQAServiceServicer(object):
    """Main QA service definition
    """

    def GenerateDraftAnswer(self, request, context):
        """Generate a draft answer for medical questions
        """
        context.set_code(grpc.StatusCode.UNIMPLEMENTED)
        context.set_details('Method not implemented!')
        raise NotImplementedError('Method not implemented!')


def add_MedicalQAServiceServicer_to_server(servicer, server):
    rpc_method_handlers = {
            'GenerateDraftAnswer': grpc.unary_unary_rpc_method_handler(
                    servicer.GenerateDraftAnswer,
                    request_deserializer=medical__service__pb2.QuestionRequest.FromString,
                    response_serializer=medical__service__pb2.QuestionResponse.SerializeToString,
            ),
    }
    generic_handler = grpc.method_handlers_generic_handler(
            'backend.MedicalQAService', rpc_method_handlers)
    server.add_generic_rpc_handlers((generic_handler,))


 # This class is part of an EXPERIMENTAL API.
class MedicalQAService(object):
    """Main QA service definition
    """

    @staticmethod
    def GenerateDraftAnswer(request,
            target,
            options=(),
            channel_credentials=None,
            call_credentials=None,
            insecure=False,
            compression=None,
            wait_for_ready=None,
            timeout=None,
            metadata=None):
        return grpc.experimental.unary_unary(request, target, '/backend.MedicalQAService/GenerateDraftAnswer',
            medical__service__pb2.QuestionRequest.SerializeToString,
            medical__service__pb2.QuestionResponse.FromString,
            options, channel_credentials,
            insecure, call_credentials, compression, wait_for_ready, timeout, metadata)
