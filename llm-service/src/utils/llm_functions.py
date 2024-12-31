from enum import Enum
from typing import Literal

# Define the function for OpenAI
class QuestionCategory(Enum):
    MEDICAL = "medical"
    TRANSPORT = "transport"
    SCHEDULE = "schedule"
    PACE_CENTER = "pace_center"
    TAXONOMY = "taxonomy"

category_function = {
    "name": "categorize_question",
    "description": "Categorize the patient's question into one specific category",
    "parameters": {
        "type": "object",
        "properties": {
            "category": {
                "type": "string",
                "enum": [cat.value for cat in QuestionCategory],
                "description": "The category that best matches the question"
            }
        },
        "required": ["category"]
    }
}