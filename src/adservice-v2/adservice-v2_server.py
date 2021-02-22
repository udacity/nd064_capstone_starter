#!/usr/bin/python

import os
import random
import time
import traceback
from concurrent import futures

import grpc

from grpc_health.v1 import health_pb2
from grpc_health.v1 import health_pb2_grpc

from logger import getJSONLogger
logger = getJSONLogger('adservice-v2-server')


class AdServiceV2():
    # TODO:
    # Implemet the Ad service business logic

    # Uncomment to enable the HealthChecks for the Ad service
    # Note: These are needed for the liveness and readiness probes
    # def Check(self, request, context):
    #     return health_pb2.HealthCheckResponse(
    #         status=health_pb2.HealthCheckResponse.SERVING)
    #
    # def Watch(self, request, context):
    #     return health_pb2.HealthCheckResponse(
    #         status=health_pb2.HealthCheckResponse.UNIMPLEMENTED)


if __name__ == "__main__":
    logger.info("initializing adservice-v2")

    # TODO:
    # create gRPC server, add the Ad-v2 service and start it

    # Uncomment to add the HealthChecks to the gRPC server to the Ad-v2 service
    # health_pb2_grpc.add_HealthServicer_to_server(service, server)
