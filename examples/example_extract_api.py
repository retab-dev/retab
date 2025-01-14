from uiform import UiForm

uiclient = UiForm()

# Or use our all-in-one API
# Returns a standard OpenAI response
response = uiclient.documents.extractions.parse(
    document = "freight/booking_confirmation.jpg", 
    model="gpt-4o-mini",
    json_schema = {
      'X-SystemPrompt': 'You are a useful assistant.',
      'properties': {
          'name': {
              'X-FieldPrompt': 'Provide a descriptive and concise name for the event.',
              'description': 'The name of the calendar event.',
              'title': 'Name',
              'type': 'string'
          },
          'date': {
              'X-ReasoningPrompt': 'The user can mention it in any format, like **next week** or **tomorrow**. Infer the right date format from the user input.',
              'description': 'The date of the calendar event in ISO 8601 format.',
              'title': 'Date',
              'type': 'string'
          }
      },
      'required': ['name', 'date'],
      'title': 'CalendarEvent',
      'type': 'object'
  },
   modality="text"
)

print(response.choices[0].message.content)



