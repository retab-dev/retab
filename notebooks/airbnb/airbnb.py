from pydantic import BaseModel, Field, ConfigDict
from typing import List
from enum import Enum, IntEnum

system_prompt = """
Your task is to extract structured information from pitch decks based on the provided schema. Carefully analyze the content to identify relevant details and populate each field accurately while maintaining consistency across extracted data.

Key Guidelines:
	1.	Understanding the Schema:
	•	The schema includes key entities such as the company, its investors, competitors, clients, team members, and business model. Each field must be carefully reviewed and filled with precise data extracted from the deck.
	•	Relationships between companies (e.g., investor, competitor, client) should be determined based on contextual evidence within the deck.
	2.	Contextual Analysis:
	•	Consider multiple slides to understand the startup's positioning, stakeholders, and market landscape.
	•	Extract information by recognizing patterns, such as recurring names or mentions in key slides (e.g., cap tables for investors, competitive analysis for competitors, and traction slides for clients).
	3.	Consistency and Accuracy:
	•	Ensure extracted data is consistent across all fields and accurately reflects the content of the deck.
	•	Avoid duplication or conflicting information by cross-referencing details from various slides.
	4.	Data Formatting:
	•	Company Name: Extract the official name as presented in the deck, ensuring correctness.
	•	Website: Extract and format URLs correctly 
	•	Country: Convert to the appropriate two-letter country code for uniformity.
	•	Business Model: Determine whether the startup follows B2B, B2C, or C2C based on their product offerings and customer base.
	5.	Handling Team Data:
	•	Identify key team members, ensuring correct extraction of their first name, last name, title, and previous experiences.
	•	Prioritize executives and founding members when available.
	6.	Dealing with Missing Information:
	•	If some details are not explicitly stated, infer them based on available context or leave them blank if they cannot be determined with certainty.
	•	Ensure no assumptions are made beyond what is supported by the content.

Extraction Output:
	•	The extracted data should strictly adhere to the defined schema, ensuring each field is correctly populated according to its description.
	•	The output should be structured, with proper data types (e.g., strings, enumerations, lists) and validated before final submission.

By following these principles, the extracted information will be comprehensive, accurate, and aligned with the expected data structure for further processing and analysis.
"""

class CompanyEnum(str, Enum):
    school = 'school'
    investor = 'investor'
    startup = 'startup'
    corporate = 'corporate'

class CompanyRelation(str, Enum):
    founderBackground = 'founderBackground'
    investor = 'investor'
    competitor = 'competitor'
    client = 'client'
    partnership = 'partnership'

class Company(BaseModel):
    name: str = Field(...,
        description="Name of the identified company",
        json_schema_extra={
            "X-FieldPrompt": "Look for the name of the company, or derive it from the logo"
        }
    )
    type: CompanyEnum = Field(...,
        description="Type of the identified company",
        json_schema_extra={
            "X-FieldPrompt": "Guess the type depending on slide context"
        }
    )
    relationship: CompanyRelation = Field(...,
        description="Relationship of the identified company with the startup from the deck",
        json_schema_extra={
            "X-FieldPrompt": "Guess the relationship of the identified company with the startup from the deck"
        }
    )

class BusinessModel(str, Enum):
    B2B = 'B2B'
    B2C = 'B2C'
    C2C = 'C2C'

class TeamMember(BaseModel):
    firstName: str = Field(...,
        description="First name of the team member",
        json_schema_extra={
            "X-FieldPrompt": "Look for the first name of the team member"
        }
    )
    lastName: str = Field(...,
        description="Last name of the team member",
        json_schema_extra={
            "X-FieldPrompt": "Look for the last name of the team member"
        }
    )
    title: str = Field(...,
        description="Title, position of the team member",
        json_schema_extra={
            "X-FieldPrompt": "Look for the title or position of the team member"
        }
    )
    pastExperiences: str = Field(...,
        description="Previous experiences",
        json_schema_extra={
            "X-FieldPrompt": "Look for previous experiences of the team member: jobs, education, companies"
        }
    )

class Deck(BaseModel):
    model_config = ConfigDict(
        json_schema_extra = {
            "X-SystemPrompt": system_prompt
        }
    )
    name: str = Field(...,
        description="Name of the company",
        json_schema_extra={
            "X-FieldPrompt": "Look for the name of the company"
        }
    )
    website: str = Field(...,
        description="URL of the company website",
        json_schema_extra={
            "X-FieldPrompt": "Look for the company URL in the deck. It is often specified in the founder's email. make sure to format as domainname.domainextesion , without www or http"
        }
    )
    country: str = Field(...,
        description="2 letter Country code of the company ",
        json_schema_extra={
            "X-FieldPrompt": "Look for the companys HQ location, in 2 letter format"
        }
    )
    investors: Company = Field(...,
        description="Existing investors in the startup",
        json_schema_extra={
            "X-FieldPrompt": "Look for investors, often cited in slides with backed by or cap tables",
            "X-ReasoningPrompt": "Think about the investors. Detail your thought process, and explain logically, step by step, who are the investors, detailing the cues and evidence for each one. For example: 1) Found X in cap table showing Y% ownership 2) Identified Z from 'backed by' slide with logo and investment date 3) etc."
        }
    )
    competitors: Company = Field(...,
        description="Competitors identified by the startup in the deck",
        json_schema_extra={
            "X-FieldPrompt": "Look for competitors, often cited in competition slides",
            "X-ReasoningPrompt": "Walk through your analysis of the competitive landscape step by step: 1) Which companies are positioned as direct competitors? 2) What specific features/offerings overlap? 3) What market segments do they compete in? 4) What evidence supports each competitor identification?"
        }
    )
    clients: Company = Field(...,
        description="Clients (B2B) in pipe or signed by the startup in the deck",
        json_schema_extra={
            "X-FieldPrompt": "Look for clients, often cited in traction or commercial pipeline slides slides",
            "X-ReasoningPrompt": "Break down your client analysis systematically: 1) Which companies are mentioned as current clients vs pipeline? 2) What evidence shows their client status (testimonials, case studies, logos)? 3) What stage is each client relationship? 4) Are there any specific metrics or success stories mentioned?"
        }
    )
    team: List[TeamMember] = Field(...,
        description="Company Team",
        json_schema_extra={
            "X-FieldPrompt": "Look for the founding team and executives at the company",
            "X-ReasoningPrompt": "Analyze the team composition methodically: 1) Who are the key founders/executives? 2) What specific evidence validates their roles? 3) How did you verify their background and experience? 4) What makes their experience relevant to this venture?"
        }
    )
    businessModel: BusinessModel = Field(...,
        description="Business model of the startup",
        json_schema_extra={
            "X-FieldPrompt": "Guess the business model based on products, offering, clients",
            "X-ReasoningPrompt": "Explain your business model classification process: 1) Who are the primary customers? 2) What is the revenue generation method? 3) How does the product/service flow between parties? 4) What specific slides or content supports this classification?"
        }
    )

