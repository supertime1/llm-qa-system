import grpc
from concurrent import futures
import logging
import yaml
import os

from . import medical_qa_pb2
from . import medical_qa_pb2_grpc
from .services.llm_service import LLMService

class MedicalQAService(medical_qa_pb2_grpc.MedicalQAServiceServicer):
    def __init__(self, config_path: str = None):
        if config_path is None:
            config_path = os.path.join(os.path.dirname(__file__), '../config/config.yaml')
        
        with open(config_path, 'r') as f:
            self.config = yaml.safe_load(f)
        
        self.llm_service = LLMService(self.config)
        self.logger = logging.getLogger(__name__)

    async def GenerateDraftAnswer(self, request, context):
        try:
            self.logger.info(f"Received question request: {request.question_id}")
            
            answer, confidence_score, references = await self.llm_service.generate_answer(
                request.question_text,
                request.user_context
            )

            self.logger.info(f"Generated answer for question: {request.question_id}")
            
            # Ensure all fields are properly set
            response = medical_qa_pb2.QuestionResponse(
                question_id=request.question_id,
                draft_answer=str(answer),  # Ensure it's a string
                confidence_score=float(confidence_score),  # Ensure it's a float
                references=list(references)  # Ensure it's a list of strings
            )
            
            return response

        except Exception as e:
            self.logger.error(f"Error processing question {request.question_id}: {str(e)}")
            context.set_code(grpc.StatusCode.INTERNAL)
            context.set_details(f'Error generating draft: {str(e)}')
            # Return empty response with required fields
            return medical_qa_pb2.QuestionResponse(
                question_id=request.question_id,
                draft_answer="",
                confidence_score=0.0
            )

def serve():
    # Load config for server
    config_path = os.path.join(os.path.dirname(__file__), '../config/config.yaml')
    with open(config_path, 'r') as f:
        config = yaml.safe_load(f)

    server = grpc.aio.server(
        futures.ThreadPoolExecutor(max_workers=config['server']['max_workers'])
    )
    medical_qa_pb2_grpc.add_MedicalQAServiceServicer_to_server(
        MedicalQAService(), server
    )
    server.add_insecure_port(f'[::]:{config["server"]["port"]}')
    
    return server

async def main():
    server = serve()
    await server.start()
    logging.info("LLM Service started on port 50051")
    await server.wait_for_termination()

if __name__ == '__main__':
    logging.basicConfig(
        level=logging.INFO,
        format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
    )
    
    import asyncio
    asyncio.run(main())