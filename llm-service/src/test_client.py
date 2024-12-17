import grpc
import logging
from . import medical_qa_pb2
from . import medical_qa_pb2_grpc

def run():
    with grpc.insecure_channel('localhost:50051') as channel:
        stub = medical_qa_pb2_grpc.MedicalQAServiceStub(channel)
        
        # Create a test request
        request = medical_qa_pb2.QuestionRequest(
            question_id="test-123",
            question_text="What are the symptoms of COVID-19?"
        )
        
        try:
            logging.info("Sending request to server...")
            response = stub.GenerateDraftAnswer(request)
            logging.info("Server response received:")
            logging.info(f"Question ID: {response.question_id}")
            logging.info(f"Draft Answer: {response.draft_answer}")
            logging.info(f"Confidence Score: {response.confidence_score}")
            logging.info(f"References: {response.references}")
        except grpc.RpcError as e:
            logging.error(f"RPC failed: {e}")

if __name__ == '__main__':
    logging.basicConfig(level=logging.INFO)
    run()