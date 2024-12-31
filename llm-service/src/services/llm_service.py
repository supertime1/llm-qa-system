from openai import AsyncOpenAI
from typing import Dict, Tuple
import json
from ..utils.prompt_builder import PromptBuilder
from ..medical_service_pb2 import UserContext
from ..utils.llm_functions import category_function
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

    async def triage_question(self, question: str) -> str:
        """
        Triage the question to determine if it is a medical question, transport question, schedule question, general questions regarding the PACE center, or a simple taxonomy.
        Args:
            question: The question text
        Returns:
            str: The category of the question
        """
        try:
            messages = [
                {"role": "system", "content": "You are a PACE center AI assitant that categorizes member questions."},
                {"role": "user", "content": question}
            ]

            response = await self.client.chat.completions.create(
                model=self.config['model'],
                messages=messages,
                functions=[category_function],
                function_call={"name": "categorize_question"}
            )

            category = json.loads(response.choices[0].message.function_call.arguments)['category']
            print(f"Category returned by LLM: {category}")
            return category
        
        except Exception as e:
            self.logger.error(f"Error processing question {question}: {str(e)}")
            return "taxonomy"



    # TODO: Add a function to generate a response based on the category
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