# Generated by the gRPC Python protocol compiler plugin. DO NOT EDIT!
"""Client and server classes corresponding to protobuf-defined services."""
import grpc

import classifier_pb2 as classifier__pb2


class ClassifierStub(object):
    """Missing associated documentation comment in .proto file."""

    def __init__(self, channel):
        """Constructor.

        Args:
            channel: A grpc.Channel.
        """
        self.ClassifyEmail = channel.unary_unary(
                '/classifier.Classifier/ClassifyEmail',
                request_serializer=classifier__pb2.ClassifyRequest.SerializeToString,
                response_deserializer=classifier__pb2.ClassifyResponse.FromString,
                )


class ClassifierServicer(object):
    """Missing associated documentation comment in .proto file."""

    def ClassifyEmail(self, request, context):
        """Missing associated documentation comment in .proto file."""
        context.set_code(grpc.StatusCode.UNIMPLEMENTED)
        context.set_details('Method not implemented!')
        raise NotImplementedError('Method not implemented!')


def add_ClassifierServicer_to_server(servicer, server):
    rpc_method_handlers = {
            'ClassifyEmail': grpc.unary_unary_rpc_method_handler(
                    servicer.ClassifyEmail,
                    request_deserializer=classifier__pb2.ClassifyRequest.FromString,
                    response_serializer=classifier__pb2.ClassifyResponse.SerializeToString,
            ),
    }
    generic_handler = grpc.method_handlers_generic_handler(
            'classifier.Classifier', rpc_method_handlers)
    server.add_generic_rpc_handlers((generic_handler,))


 # This class is part of an EXPERIMENTAL API.
class Classifier(object):
    """Missing associated documentation comment in .proto file."""

    @staticmethod
    def ClassifyEmail(request,
            target,
            options=(),
            channel_credentials=None,
            call_credentials=None,
            insecure=False,
            compression=None,
            wait_for_ready=None,
            timeout=None,
            metadata=None):
        return grpc.experimental.unary_unary(request, target, '/classifier.Classifier/ClassifyEmail',
            classifier__pb2.ClassifyRequest.SerializeToString,
            classifier__pb2.ClassifyResponse.FromString,
            options, channel_credentials,
            insecure, call_credentials, compression, wait_for_ready, timeout, metadata)
