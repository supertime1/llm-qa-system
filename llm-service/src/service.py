import grpc
from concurrent import futures
import logging
from typing import Optional

# This will be generated after running the protoc command
import medical_qa_pb2
import medical_qa_pb2_grpc

class MedicalQAService(medical_qa_pb2_grpc.MedicalQAServiceServicer):
    def __init__(self):
        # Initialize LLM model here
        self.model = None
        self.tokenizer = None

    def GenerateDraftAnswer(self, request, context):
        try:
            # TODO: Implement LLM logic
            draft_answer = "This is a placeholder draft answer"
            
            return medical_qa_pb2.AnswerResponse(
                question_id=request.question_id,
                draft_answer=draft_answer,
                confidence_score=0.95,
                references=["Reference 1", "Reference 2"]
            )
        except Exception as e:
            context.set_code(grpc.StatusCode.INTERNAL)
            context.set_details(f'Error generating draft: {str(e)}')
            return medical_qa_pb2.AnswerResponse()

def serve():
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
    medical_qa_pb2_grpc.add_MedicalQAServiceServicer_to_server(
        MedicalQAService(), server
    )
    server.add_insecure_port('[::]:50051')
    server.start()
    server.wait_for_termination()

if __name__ == '__main__':
    logging.basicConfig(level=logging.INFO)
    serve()