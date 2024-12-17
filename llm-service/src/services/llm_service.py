from openai import AsyncOpenAI
from typing import Dict, Tuple
from ..utils.prompt_builder import PromptBuilder
from ..medical_qa_pb2 import UserContext

class LLMService:
    def __init__(self, config: Dict):
        """
        Initialize LLM service with configuration
        Args:
            config: Dictionary containing OpenAI configuration
        """
        self.client = AsyncOpenAI(api_key=config['openai']['api_key'])
        self.config = config['openai']
        self.prompt_builder = PromptBuilder()

    async def generate_answer(self, question: str, user_context: UserContext) -> Tuple[str, float, list]:
        """
        Generate an answer using the OpenAI API
        Args:
            question: The question text
            user_context: UserContext protobuf message containing patient information
        Returns: 
            Tuple[str, float, list]: (answer, confidence_score, references)
        """
        try:
            messages = [
                {
                    "role": "system",
                    "content": self.prompt_builder.build_system_prompt(user_context)
                },
                {
                    "role": "user",
                    "content": f"Question: {question}\n\nPlease provide a detailed medical response, including any relevant references."
                }
            ]

            response = await self.client.chat.completions.create(
                model=self.config['model'],
                messages=messages,
                temperature=self.config['temperature'],
                max_tokens=self.config['max_tokens']
            )

            answer = str(response.choices[0].message.content)
            confidence_score = float(0.95 if response.choices[0].finish_reason == "stop" else 0.5)
            references = list(self._extract_references(answer))

            return answer, confidence_score, references

        except Exception as e:
            # Return empty but valid response
            return "", 0.0, []

    def _extract_references(self, answer: str) -> list:
        """
        Extract references from the answer text
        Args:
            answer: The generated answer text
        Returns:
            list: List of extracted references
        """
        references = []
        for line in answer.split('\n'):
            if line.strip().startswith(('ref:', 'reference:', '[', '1.', '2.')):
                references.append(line.strip())
        return references