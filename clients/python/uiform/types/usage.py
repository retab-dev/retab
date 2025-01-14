from pydantic import BaseModel

# Monthly Usage
class MonthlyUsageResponseContent(BaseModel):
    request_count: int

MonthlyUsageResponse = MonthlyUsageResponseContent