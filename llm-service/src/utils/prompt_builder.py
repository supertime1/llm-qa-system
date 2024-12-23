from datetime import datetime
from typing import List
from google.protobuf.timestamp_pb2 import Timestamp
from ..medical_service_pb2 import UserContext, BiometricData, ChatMessage


class PromptBuilder:
    @staticmethod
    def build_system_prompt(user_context: UserContext) -> str:
        """
        Build system prompt from UserContext protobuf message
        """
        return f"""You are a medical AI assistant helping medical professionals such as doctors, nurses, and caregiving professionals to draft answers for medical and caregiving related questions. Please provide accurate medical information based on the following patient context:

Patient Information:
- Age: {user_context.user_info.age}
- Gender: {user_context.user_info.gender}
- Medical History: {', '.join(user_context.user_info.medical_history)}

Recent Biometric Data:
{PromptBuilder._format_biometrics(user_context.biometric_data)}

Recent Chat History:
{PromptBuilder._format_chat_history(user_context.chat_history)}

Please provide a clear, professional response that a medical professional can review."""

    @staticmethod
    def _format_biometrics(biometrics: List[BiometricData]) -> str:
        """
        Format BiometricData protobuf messages
        Each BiometricData has:
        - type: str
        - value: str
        - timestamp: google.protobuf.Timestamp
        """
        formatted = []
        for b in biometrics:
            # Convert protobuf Timestamp to datetime
            dt = datetime.fromtimestamp(b.timestamp.seconds + b.timestamp.nanos/1e9)
            formatted.append(f"- {b.type}: {b.value} (recorded: {dt.strftime('%Y-%m-%d %H:%M:%S')})")
        return "\n".join(formatted)

    @staticmethod
    def _format_chat_history(history: List[ChatMessage]) -> str:
        """
        Format ChatMessage protobuf messages
        Each ChatMessage has:
        - role: str
        - content: str
        - timestamp: google.protobuf.Timestamp
        """
        formatted = []
        for msg in history:
            # Convert protobuf Timestamp to datetime
            dt = datetime.fromtimestamp(msg.timestamp.seconds + msg.timestamp.nanos/1e9)
            formatted.append(f"- {msg.role}: {msg.content} ({dt.strftime('%Y-%m-%d %H:%M:%S')})")
        return "\n".join(formatted)