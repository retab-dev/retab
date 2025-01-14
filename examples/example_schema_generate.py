from uiform import UiForm

uiclient = UiForm()

schema_obj = uiclient.schemas.generate(
    documents = [
        "freight/booking_confirmation_1.jpg",
        "freight/booking_confirmation_2.jpg"
    ]
)