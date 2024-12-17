import grpc
import logging
from datetime import datetime
from google.protobuf.timestamp_pb2 import Timestamp
from . import medical_qa_pb2
from . import medical_qa_pb2_grpc

def create_test_request() -> medical_qa_pb2.QuestionRequest:
    """Create a test request with sample data"""
    
    # Create UserInfo
    user_info = medical_qa_pb2.UserInfo(
        age="45",
        gender="Female",
        medical_history=[
            "Type 2 Diabetes",
            "Hypertension",
            "Asthma"
        ]
    )

    # Create some sample biometric data
    current_time = Timestamp()
    current_time.GetCurrentTime()  # Sets to current time

    biometric_data = [
        medical_qa_pb2.BiometricData(
            type="Blood Pressure",
            value="120/80 mmHg",
            timestamp=current_time
        ),
        medical_qa_pb2.BiometricData(
            type="Blood Sugar",
            value="140 mg/dL",
            timestamp=current_time
        ),
        medical_qa_pb2.BiometricData(
            type="Heart Rate",
            value="72 bpm",
            timestamp=current_time
        )
    ]

    # Create some sample chat history
    chat_history = [
        medical_qa_pb2.ChatMessage(
            role="user",
            content="I've been experiencing increased thirst lately.",
            timestamp=current_time
        ),
        medical_qa_pb2.ChatMessage(
            role="assistant",
            content="This could be related to your diabetes. Let's discuss this with your doctor.",
            timestamp=current_time
        )
    ]

    # Create UserContext
    user_context = medical_qa_pb2.UserContext(
        user_info=user_info,
        biometric_data=biometric_data,
        chat_history=chat_history
    )

    # Create the final request
    return medical_qa_pb2.QuestionRequest(
        question_id="test-123",
        question_text="What are the potential complications of uncontrolled diabetes, and what preventive measures should I take?",
        user_context=user_context
    )

async def run():
    async with grpc.aio.insecure_channel('localhost:50051') as channel:
        stub = medical_qa_pb2_grpc.MedicalQAServiceStub(channel)
        
        try:
            # Create and send request
            request = create_test_request()
            logging.info("Sending request to server...")
            logging.info(f"Question: {request.question_text}")
            
            response = await stub.GenerateDraftAnswer(request)
            
            # Log the response details
            logging.info("\nServer response received:")
            logging.info(f"Question ID: {response.question_id}")
            logging.info("\nDraft Answer:")
            logging.info(f"{response.draft_answer}")
            logging.info(f"\nConfidence Score: {response.confidence_score}")
            
            if response.references:
                logging.info("\nReferences:")
                for ref in response.references:
                    logging.info(f"- {ref}")
            else:
                logging.info("\nNo references provided")

        except grpc.RpcError as e:
            logging.error(f"RPC failed: {e.code()}")
            logging.error(f"Details: {e.details()}")

if __name__ == '__main__':
    logging.basicConfig(
        level=logging.INFO,
        format='%(asctime)s - %(levelname)s - %(message)s'
    )
    
    import asyncio
    asyncio.run(run())