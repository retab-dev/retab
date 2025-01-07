# uiform

UiForm is a **modern**, **flexible**, and **AI-powered** document processing API that helps you:

- Create from JSON schemas and Pydantic models with zero boilerplate
- Add AI capabilities for automated document processing, that is compatible with any data structure
- Create annotated datasets to distill or finetune your models

Our goal is to make the process of analyzing documents and unstructured data as **easy** and **transparent** as possible. You come with your own API key from your favorite AI provider, and we handle the rest.

## Quickstart

<Steps>
  <Step title="Load a JSON Schema and some Documents">
    Setup your API keys and load a JSON Schema and some Documents.
  </Step>
  <Step title="Extract data from your documents with our Python SDK">
    Use UiForm to extract data from your documents
  </Step>
</Steps>





### 1- Load a JSON Schema and some Documents

Save this [JSON Schema](https://github.com/UiForm/uiform/blob/main/notebooks/freight/booking_confirmation_json_schema.json) as `json_schema.json`

Download this [example document](https://github.com/UiForm/uiform/blob/main/notebooks/freight/booking_confirmation.jpg) as `example.jpg`

<CodeGroup>
```python main.py
import json

with open("booking_confirmation_json_schema.json", "r") as f:
    json_schema = json.load(f)

document = "booking_confirmation.jpg"
```

```json json_schema.json
{
  "title": "RoadBookingConfirmationData",
  "type": "object",
  "properties": {
    "thinking": {
      "anyOf": [
        {
          "type": "string"
        },
        {
          "type": "null"
        }
      ],
      "default": null,
      "description": "Document your analysis of the overall booking confirmation. Include:\n\n1. Document type identification and confidence\n2. Key parties involved (client, carrier, etc.)\n3. Overall structure of the shipments\n4. Critical booking details and requirements\n5. Any ambiguities or unclear elements\n\nExample: \"Analysis shows:\n- Clear booking confirmation document from ABC Corp\n- Contains 2 distinct shipments under single booking reference\n- Client identified as ABC Corp from letterhead\n- Total value and payment terms clearly stated\n- All required signatures present\"\n\nExplain your reasoning for key decisions made in interpreting the document structure.",
      "title": "Thinking"
    },
    "booking_id": {
      "anyOf": [
        {
          "type": "string"
        },
        {
          "type": "null"
        }
      ],
      "default": null,
      "description": "Unique identifier for the booking used for billing reference",
      "title": "Booking Id"
    },
    "payment": {
      "$ref": "#/$defs/TransportPriceData",
      "description": "Payment data"
    },
    "client": {
      "$ref": "#/$defs/ClientData",
      "description": "Client data used for billing"
    },
    "number_of_shipments": {
      "anyOf": [
        {
          "type": "integer"
        },
        {
          "type": "null"
        }
      ],
      "default": null,
      "description": "Number of shipments in the booking. Most frequently 1, but can be 2,3 to 10-20",
      "title": "Number Of Shipments"
    },
    "shipments_thinking": {
      "anyOf": [
        {
          "type": "string"
        },
        {
          "type": "null"
        }
      ],
      "default": null,
      "description": "Document your analysis for each shipment. For each shipment identified in the document, include:\n\n1. A detailed description of the goods being transported, including any reference numbers, dangerous goods information, and packing details.\n2. The sender\u2019s information, including the company name, address, contact details, and any relevant observations about the pickup process.\n3. The recipient\u2019s information, including the company name, address, contact details, and any relevant observations about the delivery process.\n4. Any trucking constraints, including vehicle requirements, loading/unloading equipment, and specific transport conditions.\n\nExample:\n\"Shipment Analysis:\n1: SSP BOC - 4 EUR pallets, 10ML, 7,623 kg  (UN 3082, liquid, 9, GE III)\n2: AMC(31) BOS 2249 - 8 pallets, 7ML, 3,240 kg\n3: VETISOL BOS 2195 - 10 pallets, 9ML, 3,780 kg\n4: VETISOL BOS 2282 - 6 pallets, 5ML, 2,020 kg (UN 3077, solid, 9, GE III)\n5: SOERMEL BOS 2264 - 2 pallets, 8ML, 1,300 kg (UN 3077, solid, 9, GE III)\n6: DUMOULIN BOS 2281 - 15 pallets, 11ML, 10,400 kg\n7: SOLUSTIL BOS 1764 - 20 pallets, 12ML, 24,800 kg Refrigerated vehicle (-20\u00b0C)\n8: TECMA TAR 323 - 1 pallet, 2ML, 1,475 kg\n9: PAILLET BOS 2290 - 8 pallets, 10ML, 4,945 kg\n\nEnsure that each shipment is clearly detailed with all relevant information. Do not omit any critical details.\nMost frequently, there is going to be one or two shipments in the document, but sometimes you can have a massive number of shipments on several lines, more like a table containing a long list of shipments.\nIn this case, you should create a shipment for each line of the table, it can go to 10-20 shipments sometimes.\n",
      "title": "Shipments Thinking"
    },
    "shipments": {
      "description": "List of shipment data. The number of entries should correspond to the total loading/delivery point combinations identified in the document.",
      "items": {
        "$ref": "#/$defs/ShipmentData"
      },
      "title": "Shipments",
      "type": "array"
    }
  },
  "required": [
    "payment",
    "client",
    "shipments"
  ],
  "$defs": {
    "AddressDataSimple": {
      "type": "object",
      "properties": {
        "city": {
          "anyOf": [
            {
              "type": "string"
            },
            {
              "type": "null"
            }
          ],
          "default": null,
          "description": "City, district, suburb, town, or village.",
          "title": "City"
        },
        "postal_code": {
          "anyOf": [
            {
              "type": "string"
            },
            {
              "type": "null"
            }
          ],
          "default": null,
          "description": "ZIP or postal code. If french postal code, it should be a pure number, without letters.",
          "title": "Postal Code"
        },
        "country": {
          "anyOf": [
            {
              "type": "string"
            },
            {
              "type": "null"
            }
          ],
          "default": null,
          "description": "Two-letter country code (ISO 3166-1 alpha-2).",
          "title": "Country"
        },
        "line1": {
          "anyOf": [
            {
              "type": "string"
            },
            {
              "type": "null"
            }
          ],
          "default": null,
          "description": "Address line 1 (e.g., street, PO Box, or company name).",
          "title": "Line1"
        },
        "line2": {
          "anyOf": [
            {
              "type": "string"
            },
            {
              "type": "null"
            }
          ],
          "default": null,
          "description": "Address line 2 (e.g., apartment, suite, unit, or building).",
          "title": "Line2"
        }
      },
      "title": "AddressDataSimple"
    },
    "ClientData": {
      "type": "object",
      "properties": {
        "company_name": {
          "anyOf": [
            {
              "type": "string"
            },
            {
              "type": "null"
            }
          ],
          "default": null,
          "description": "Legal company name for billing",
          "title": "Company Name"
        },
        "SIREN": {
          "anyOf": [
            {
              "type": "string"
            },
            {
              "type": "null"
            }
          ],
          "default": null,
          "description": "SIREN number for billing and tax purposes. Usually located at the top or the very bottom of the document. (small text), and located alongside the other billing informations",
          "title": "Siren"
        },
        "VAT_number_thinking": {
          "anyOf": [
            {
              "type": "string"
            },
            {
              "type": "null"
            }
          ],
          "default": null,
          "description": "The VAT number is present for billing and tax purposes. Usually located at the top or the very bottom of the document, and located alongside the other billing informations (small text).\nIn France the VAT number format is FR {2 digits validation key} {first 9 digits SIREN number} so try use the SIREN number if available to help validate the VAT number.\nDocument your reasoning for the VAT number identification:\n1. Look for numbers prefixed with 'FR' or labeled as 'TVA', 'VAT', or 'n\u00b0 TVA'\n2. Verify the format matches FR followed by key of 2 digits and 9 digits of SIREN number\n3. Note where you found it in the document\n\nExample: \"Found VAT number FR12345678901 in footer next to SIREN number\"\n\nIf multiple possible VAT numbers exist, explain your choice. If none found, state \"No VAT number identified in document.",
          "title": "Vat Number Thinking"
        },
        "VAT_number": {
          "anyOf": [
            {
              "type": "string"
            },
            {
              "type": "null"
            }
          ],
          "default": null,
          "description": "VAT number Value",
          "title": "Vat Number"
        },
        "city": {
          "anyOf": [
            {
              "type": "string"
            },
            {
              "type": "null"
            }
          ],
          "default": null,
          "description": "City, district, suburb, town, or village of billing address",
          "title": "City"
        },
        "postal_code": {
          "anyOf": [
            {
              "type": "string"
            },
            {
              "type": "null"
            }
          ],
          "default": null,
          "description": "ZIP or postal code of billing address. If french postal code, it should be a pure number, without letters.",
          "title": "Postal Code"
        },
        "country": {
          "anyOf": [
            {
              "type": "string"
            },
            {
              "type": "null"
            }
          ],
          "default": null,
          "description": "Two-letter country code (ISO 3166-1 alpha-2).",
          "title": "Country"
        },
        "code": {
          "anyOf": [
            {
              "type": "string"
            },
            {
              "type": "null"
            }
          ],
          "default": null,
          "description": "Unique code for the client used to connect the booking to the client in the TMS. It is rarely provided. If not provided, set to None",
          "title": "Code"
        },
        "email": {
          "anyOf": [
            {
              "type": "string"
            },
            {
              "type": "null"
            }
          ],
          "default": null,
          "description": "Client email address",
          "title": "Email"
        }
      },
      "title": "ClientData"
    },
    "DangerousGoodsInfoData": {
      "type": "object",
      "properties": {
        "thinking": {
          "anyOf": [
            {
              "type": "string"
            },
            {
              "type": "null"
            }
          ],
          "default": null,
          "description": "Document your analysis of the goods information from the source document. Include:\n\n1. Description of what's being transported\n2. Any reference numbers or codes found\n3. Special handling requirements\n4. Equipment needs\n5. Safety considerations\n6. Documentation requirements\n7. The presence of returnable Euro pallets (it has to be explicitly stated that the pallets are EU pallets in the source document in the sender section (not in the legal mentions))\n\n*Some observations you may consider about the weight:*\n- Sometimes the weight appears in tons, or kilograms. If the unit is not specified, assume it is in kilograms. However, the weight may appear like \"12T400\" which means 12.4 tons.\n- Some weight examples: \"1500\" -> 1500 kg, \"1T500\" -> 1.500 tons -> 1500 kg, \"7T565\" -> 7.565 tons -> 7565 kg, \"12 tons\" -> 12000 kg, \"14 tonnes\" -> 14000 kg\n\n*Example of thinking output:*\n```\nSource indicates: \n- 12 pallets of automotive parts\n- Reference: ORDER123, BATCH456\n- Requires forklift for unloading\n- Non-stackable items\n- Fragile goods warning labels needed\"\n- Weight is in displayed in tons - its value is 5.4 tons, which is equivalent to 5400 kg.\n```\nList all relevant details found in the source document. Do not make assumptions about information that isn't explicitly stated.",
          "title": "Thinking"
        },
        "weight": {
          "anyOf": [
            {
              "type": "number"
            },
            {
              "type": "null"
            }
          ],
          "default": null,
          "description": "Weight of the dangerous good (in kilograms)",
          "title": "Weight"
        },
        "UN_code": {
          "anyOf": [
            {
              "type": "integer"
            },
            {
              "type": "null"
            }
          ],
          "default": null,
          "description": "UN code of the dangerous good",
          "title": "Un Code"
        },
        "UN_label": {
          "anyOf": [
            {
              "type": "string"
            },
            {
              "type": "null"
            }
          ],
          "default": null,
          "description": "UN label of the dangerous good",
          "title": "Un Label"
        },
        "UN_class": {
          "anyOf": [
            {
              "type": "string"
            },
            {
              "type": "null"
            }
          ],
          "default": null,
          "description": "UN class of the dangerous good",
          "title": "Un Class"
        },
        "UN_packing_group": {
          "anyOf": [
            {
              "enum": [
                "I",
                "II",
                "III"
              ],
              "type": "string"
            },
            {
              "type": "null"
            }
          ],
          "default": null,
          "description": "UN packing group of the dangerous good",
          "title": "Un Packing Group"
        },
        "ADR_tunnel_code": {
          "anyOf": [
            {
              "enum": [
                "B",
                "B1000C",
                "B/D",
                "B/E",
                "C",
                "C5000D",
                "C/D",
                "C/E",
                "D",
                "D/E",
                "E",
                "-"
              ],
              "type": "string"
            },
            {
              "type": "null"
            }
          ],
          "default": null,
          "description": "ADR tunnel code of the dangerous good",
          "title": "Adr Tunnel Code"
        }
      },
      "title": "DangerousGoodsInfoData"
    },
    "DeliveryDatetimeData": {
      "type": "object",
      "properties": {
        "date": {
          "anyOf": [
            {
              "format": "date",
              "type": "string"
            },
            {
              "type": "null"
            }
          ],
          "default": null,
          "description": "Date of the delivery. ISO 8601 Date Format: YYYY-MM-DD",
          "title": "Date"
        },
        "start_time": {
          "anyOf": [
            {
              "format": "iso-time",
              "type": "string"
            },
            {
              "type": "null"
            }
          ],
          "default": null,
          "description": "Start time of the delivery. ISO 8601 Time Format: hh:mm",
          "title": "Start Time"
        },
        "end_time": {
          "anyOf": [
            {
              "format": "iso-time",
              "type": "string"
            },
            {
              "type": "null"
            }
          ],
          "default": null,
          "description": "End time of the delivery. ISO 8601 Time Format: hh:mm. Must be greater than or equal to the start_time.",
          "title": "End Time"
        }
      },
      "title": "DeliveryDatetimeData"
    },
    "DimensionsData": {
      "type": "object",
      "properties": {
        "loading_meters_thinking": {
          "anyOf": [
            {
              "type": "string"
            },
            {
              "type": "null"
            }
          ],
          "default": null,
          "description": "A loading meter (LM) is a measurement of the cargo space occupied in a truck, calculated based on the item\u2019s footprint relative to the truck\u2019s standard internal width of 2.4 meters. This measurement helps determine how much linear floor space is needed to transport the goods efficiently.\n\nIf loading meters are provided in the source (often labeled as M\u00e8tre lin\u00e9aire (ML) or M\u00e8tre plancher (MPL)), use that value directly, noting it as \u201cUsing provided value of X loading meters.\u201d\n\nIf not, calculate as follows:\n\n\t1.\tFor each item: (length \u00d7 width) / 2.4.\n\t2.\tSum all calculated loading meters.\n\nFor example:\n\n\t\u2022\tItem 1 is 3m \u00d7 1.2m, so it takes (3 \u00d7 1.2) / 2.4 = 1.5 loading meters.\n\t\u2022\tItem 2 is 2m \u00d7 1.6m, so it takes (2 \u00d7 1.6) / 2.4 = 1.33 loading meters.\n\nTotal loading meters: 1.5 + 1.33 = 2.83 meters\n\nFor each item, write down the formula, insert values, and compute explicitly.\nTake your time to think rigourously. Do not hesitate to write down intermediate steps. WARNING: VOLUME WILL NOT BE IN ml, this is the unit used for loading meters - ml stands for metres lineaires.",
          "title": "Loading Meters Thinking"
        },
        "loading_meters": {
          "anyOf": [
            {
              "type": "number"
            },
            {
              "type": "null"
            }
          ],
          "default": null,
          "description": "Loading meters value",
          "title": "Loading Meters"
        },
        "volume_thinking": {
          "anyOf": [
            {
              "type": "string"
            },
            {
              "type": "null"
            }
          ],
          "default": null,
          "description": "Volume represents the total space occupied by an item in cubic meters, calculated as (length \u00d7 width \u00d7 height).\n\nIf volume is provided in the source, use that value directly and note it as, \u201cUsing provided value of X cubic meters.\u201d\n\nIf not, calculate as follows:\n\n\t1.\tFor each item: length \u00d7 width \u00d7 height.\n\t2.\tSum the volumes of all items for the total volume.\n\nExample:\n\n\t\u2022\tItem 1: 2m \u00d7 1m \u00d7 1.5m = 3 cubic meters\n\t\u2022\tItem 2: 1.5m \u00d7 1m \u00d7 1m = 1.5 cubic meters\n\nTotal volume: 3 + 1.5 = 4.5 cubic meters\n\nFor each item, write down the formula, insert values, and compute explicitly. Take your time to ensure accuracy, and write intermediate steps if necessary.",
          "title": "Volume Thinking"
        },
        "volume": {
          "anyOf": [
            {
              "type": "number"
            },
            {
              "type": "null"
            }
          ],
          "default": null,
          "description": "Volume value in cubic meters",
          "title": "Volume"
        }
      },
      "title": "DimensionsData"
    },
    "GoodsData": {
      "type": "object",
      "properties": {
        "thinking": {
          "anyOf": [
            {
              "type": "string"
            },
            {
              "type": "null"
            }
          ],
          "default": null,
          "description": "Document your analysis of the goods information from the source document. Include:\n\n1. Description of what's being transported\n2. Any reference numbers or codes found\n3. Special handling requirements\n4. Equipment needs\n5. Safety considerations\n6. Documentation requirements\n7. The presence of returnable Euro pallets (it has to be explicitly stated that the pallets are EU pallets in the source document in the sender section (not in the legal mentions))\n\nExample: \"Source indicates: \n- 12 pallets of automotive parts\n- Reference: ORDER123, BATCH456\n- Requires forklift for unloading\n- Non-stackable items\n- Fragile goods warning labels needed\"\n- Weight is in displayed in tons - A conversion to kg is needed\n\nList all relevant details found in the source document. Do not make assumptions about information that isn't explicitly stated.",
          "title": "Thinking"
        },
        "packing": {
          "$ref": "#/$defs/PackingData",
          "default": {
            "units": null,
            "packing_type": null,
            "supplementary_parcels": null,
            "pallets_on_ground": null,
            "number_eur_pallet": null,
            "observation": null
          },
          "description": "Packing details of the good"
        },
        "dimensions": {
          "$ref": "#/$defs/DimensionsData",
          "default": {
            "loading_meters": null,
            "volume": null
          },
          "description": "Dimensions of the good"
        },
        "weight": {
          "anyOf": [
            {
              "type": "number"
            },
            {
              "type": "null"
            }
          ],
          "default": null,
          "description": "Weight of the good (in kilograms)",
          "title": "Weight"
        },
        "temperature_infos": {
          "anyOf": [
            {
              "$ref": "#/$defs/TemperatureInfosData"
            },
            {
              "type": "null"
            }
          ],
          "default": {
            "min_temperature": null,
            "max_temperature": null,
            "category": null
          },
          "description": "Temperature infos of the good"
        },
        "dangerous_goods_infos": {
          "default": [],
          "description": "Dangerous goods infos of the good",
          "items": {
            "$ref": "#/$defs/DangerousGoodsInfoData"
          },
          "title": "Dangerous Goods Infos",
          "type": "array"
        }
      },
      "title": "GoodsData"
    },
    "PackingData": {
      "type": "object",
      "properties": {
        "units": {
          "anyOf": [
            {
              "type": "integer"
            },
            {
              "type": "null"
            }
          ],
          "default": null,
          "description": "Number of units",
          "title": "Units"
        },
        "packing_type": {
          "anyOf": [
            {
              "enum": [
                "pallet",
                "bulk",
                "container",
                "other"
              ],
              "type": "string"
            },
            {
              "type": "null"
            }
          ],
          "default": null,
          "description": "Type of packing - make sure you don't say it's a pallet (e.g. it says PAL) if you are not sure",
          "title": "Packing Type"
        },
        "supplementary_parcels": {
          "anyOf": [
            {
              "type": "integer"
            },
            {
              "type": "null"
            }
          ],
          "default": null,
          "description": "Number of additional parcels",
          "title": "Supplementary Parcels"
        },
        "pallets_on_ground_thinking": {
          "anyOf": [
            {
              "type": "string"
            },
            {
              "type": "null"
            }
          ],
          "default": null,
          "description": "This field calculates the equivalent floor area occupied by packing units, expressed in terms of standard EUR pallets (1.2m \u00d7 0.8m). Follow these steps for each pallet, then sum the results to get the total floor pallet equivalent:\n\n\t1.\tFor each pallet: (length \u00d7 width) / (1.2 \u00d7 0.8).\n\t2.\tSum the calculations to get the total floor pallet equivalent area.\n\nExample Calculations:\n\nExample 1: 6 EUR Pallets (Standard Size)\n\nIf all 6 pallets are standard EUR pallets:\n\n\t\u2022\tEach pallet occupies: (1.2 \u00d7 0.8) / (1.2 \u00d7 0.8) = 1 pallet equivalent.\n\t\u2022\tTotal for 6 pallets: 1 + 1 + 1 + 1 + 1 + 1 = 6.\n\nSo, floor pallet equivalent = 6.\n\nExample 2: Two Custom Pallets\n\n\t\u2022\tPallet 1: 1.5m \u00d7 0.9m, so 1.5 \u00d7 0.9 / (1.2 \u00d7 0.8) \u2248 1.41.\n\t\u2022\tPallet 2: 1.0m \u00d7 1.1m, so 1.0 \u00d7 1.1 / (1.2 \u00d7 0.8) \u2248 1.15.\n\nTotal floor pallet equivalent = 1.41 + 1.15 = 2.56.\n\nFor each pallet, write down the formula, insert values, and compute explicitly.\nTake your time to think rigourously. Do not hesitate to write down intermediate steps.",
          "title": "Pallets On Ground Thinking"
        },
        "pallets_on_ground": {
          "anyOf": [
            {
              "type": "number"
            },
            {
              "type": "null"
            }
          ],
          "default": null,
          "description": "pallets on ground value",
          "title": "Pallets On Ground"
        },
        "number_eur_pallet": {
          "anyOf": [
            {
              "type": "integer"
            },
            {
              "type": "null"
            }
          ],
          "default": null,
          "description": "Number of pallets that follow the European standard, i.e. marked as something like EUR or EPAL. If not explicitly mentioned, set to 0.",
          "title": "Number Eur Pallet"
        },
        "observation": {
          "anyOf": [
            {
              "type": "string"
            },
            {
              "type": "null"
            }
          ],
          "default": null,
          "description": "Information about goods, packing, and handling requirements. This includes:\n\t\u2022\tReference numbers (e.g., product codes, serial numbers)\n\t\u2022\tHandling instructions, stacking limits, and loading specs (e.g., side/rear-loading)\n\t\u2022\tVehicle, equipment, or packaging needs\n\t\u2022\tSpecial care (e.g., temperature control, fragility, safety warnings)\n\t\u2022\tLabels, placards, seal/container numbers, or other markings\nWhen you write observations, write them in the same language as the email.",
          "title": "Observation"
        }
      },
      "title": "PackingData"
    },
    "PickupDatetimeData": {
      "type": "object",
      "properties": {
        "date": {
          "anyOf": [
            {
              "format": "date",
              "type": "string"
            },
            {
              "type": "null"
            }
          ],
          "default": null,
          "description": "Date of the pickup. ISO 8601 Date Format: YYYY-MM-DD",
          "title": "Date"
        },
        "start_time": {
          "anyOf": [
            {
              "format": "iso-time",
              "type": "string"
            },
            {
              "type": "null"
            }
          ],
          "default": null,
          "description": "Start time of the pickup. ISO 8601 Time Format: hh:mm",
          "title": "Start Time"
        },
        "end_time": {
          "anyOf": [
            {
              "format": "iso-time",
              "type": "string"
            },
            {
              "type": "null"
            }
          ],
          "default": null,
          "description": "End time of the pickup. ISO 8601 Time Format: hh:mm. Must be greater than or equal to the start_time.",
          "title": "End Time"
        }
      },
      "title": "PickupDatetimeData"
    },
    "RecipientData": {
      "type": "object",
      "properties": {
        "company_name": {
          "anyOf": [
            {
              "type": "string"
            },
            {
              "type": "null"
            }
          ],
          "default": null,
          "description": "Name of the company.",
          "title": "Company Name"
        },
        "address": {
          "$ref": "#/$defs/AddressDataSimple",
          "default": {
            "city": null,
            "postal_code": null,
            "country": null,
            "line1": null,
            "line2": null
          },
          "description": "Address of the recipient."
        },
        "phone_number": {
          "anyOf": [
            {
              "type": "string"
            },
            {
              "type": "null"
            }
          ],
          "default": null,
          "description": "Phone number of the recipient.",
          "title": "Phone Number"
        },
        "email_address": {
          "anyOf": [
            {
              "format": "email",
              "type": "string"
            },
            {
              "type": "null"
            }
          ],
          "default": null,
          "description": "Email address of the recipient.",
          "title": "Email Address"
        },
        "delivery_datetime": {
          "$ref": "#/$defs/DeliveryDatetimeData",
          "default": {
            "date": null,
            "start_time": null,
            "end_time": null
          },
          "description": "delivery date and time. Must be after the pickup date and time."
        },
        "observations": {
          "anyOf": [
            {
              "type": "string"
            },
            {
              "type": "null"
            }
          ],
          "default": null,
          "description": "Include all relevant delivery information. This includes:\n\t\u2022\tReference numbers (e.g., delivery refs, order numbers)\n\t\u2022\tContact details (e.g., names, roles, extensions)\n\t\u2022\tDock/door numbers, access codes, and specific entrance instructions\n\t\u2022\tRequired documentation and any special equipment needs\n\t\u2022\tUnloading bay specs, timing preferences, and site access instructions\n\t\u2022\tParking info, unloading procedures, safety protocols, and any additional notes near recipient info\nWhen you write observations, write them in the same language as the email and the documents. If they are in French, write them in French, if they are in English, write them in English.",
          "title": "Observations"
        }
      },
      "title": "RecipientData"
    },
    "SenderData": {
      "type": "object",
      "properties": {
        "company_name": {
          "anyOf": [
            {
              "type": "string"
            },
            {
              "type": "null"
            }
          ],
          "default": null,
          "description": "Name of the company.",
          "title": "Company Name"
        },
        "address": {
          "$ref": "#/$defs/AddressDataSimple",
          "default": {
            "city": null,
            "postal_code": null,
            "country": null,
            "line1": null,
            "line2": null
          },
          "description": "Address of the sender."
        },
        "phone_number": {
          "anyOf": [
            {
              "type": "string"
            },
            {
              "type": "null"
            }
          ],
          "default": null,
          "description": "Phone number of the sender.",
          "title": "Phone Number"
        },
        "email_address": {
          "anyOf": [
            {
              "format": "email",
              "type": "string"
            },
            {
              "type": "null"
            }
          ],
          "default": null,
          "description": "Email address of the sender.",
          "title": "Email Address"
        },
        "pickup_datetime": {
          "$ref": "#/$defs/PickupDatetimeData",
          "default": {
            "date": null,
            "start_time": null,
            "end_time": null
          },
          "description": "pickup date and time."
        },
        "observations": {
          "anyOf": [
            {
              "type": "string"
            },
            {
              "type": "null"
            }
          ],
          "default": null,
          "description": "Include all relevant information for pickup and sender details. This includes:\n\t\u2022\tReference numbers (e.g., order or booking refs)\n\t\u2022\tContact details (e.g., names, roles, extensions)\n\t\u2022\tDock/door numbers, access codes, and navigation instructions\n\t\u2022\tRequired documentation and special equipment needs\n\t\u2022\tLoading bay specs, timing preferences, and site access instructions\n\t\u2022\tParking info, specific entrance locations, and any additional sender notes\nWhen you write observations, write them in the same language as the email and the documents. If they are in French, write them in French, if they are in English, write them in English.",
          "title": "Observations"
        }
      },
      "title": "SenderData"
    },
    "ShipmentData": {
      "type": "object",
      "properties": {
        "thinking": {
          "anyOf": [
            {
              "type": "string"
            },
            {
              "type": "null"
            }
          ],
          "default": null,
          "description": "Document your analysis of the shipment information. Include:\n\n1. How you identified this as a distinct shipment\n2. Key reference numbers and their sources\n3. Relationships between pickup and delivery points\n4. Any special routing or timing requirements\n5. Document references (CMR, waybills, etc.)\n\nExample: \"Source shows:\n- Single shipment identified by CMR #12345\n- Pickup from warehouse A to delivery point B\n- Must deliver within 24h window\n- References found: Order #789, Customer PO #456\"\n\nDocument your reasoning process and any key decisions made in interpreting the shipment structure.",
          "title": "Thinking"
        },
        "shipment_id": {
          "anyOf": [
            {
              "type": "string"
            },
            {
              "type": "null"
            }
          ],
          "default": null,
          "description": "Unique identifier for the shipment, most often the CMR number or CMR reference, used for tracking and documenting transport details. Can also be client / sender reference number. If several numbers are provided and the situation is ambiguous, pick the most logical one, and put other numbers in the observation field.",
          "title": "Shipment Id"
        },
        "sender": {
          "$ref": "#/$defs/SenderData",
          "default": {
            "company_name": null,
            "address": {
              "city": null,
              "country": null,
              "line1": null,
              "line2": null,
              "postal_code": null
            },
            "phone_number": null,
            "email_address": null,
            "pickup_datetime": {
              "date": null,
              "end_time": null,
              "start_time": null
            },
            "observations": null
          },
          "description": "Sender data"
        },
        "recipient": {
          "$ref": "#/$defs/RecipientData",
          "default": {
            "company_name": null,
            "address": {
              "city": null,
              "country": null,
              "line1": null,
              "line2": null,
              "postal_code": null
            },
            "phone_number": null,
            "email_address": null,
            "delivery_datetime": {
              "date": null,
              "end_time": null,
              "start_time": null
            },
            "observations": null
          },
          "description": "Recipient data"
        },
        "goods": {
          "$ref": "#/$defs/GoodsData",
          "default": {
            "packing": {
              "number_eur_pallet": null,
              "observation": null,
              "packing_type": null,
              "pallets_on_ground": null,
              "supplementary_parcels": null,
              "units": null
            },
            "dimensions": {
              "loading_meters": null,
              "volume": null
            },
            "weight": null,
            "temperature_infos": {
              "category": null,
              "max_temperature": null,
              "min_temperature": null
            },
            "dangerous_goods_infos": []
          },
          "description": "Description of transported goods"
        },
        "transport_constraints": {
          "$ref": "#/$defs/TruckData",
          "default": {
            "vehicle_type": null,
            "body_type": null,
            "tail_lift": null,
            "crane": null
          },
          "description": "List of transport constraints informations"
        }
      },
      "title": "ShipmentData"
    },
    "TemperatureInfosData": {
      "type": "object",
      "properties": {
        "min_temperature": {
          "anyOf": [
            {
              "type": "number"
            },
            {
              "type": "null"
            }
          ],
          "default": null,
          "description": "Minimum temperature (in Celsius)",
          "title": "Min Temperature"
        },
        "max_temperature": {
          "anyOf": [
            {
              "type": "number"
            },
            {
              "type": "null"
            }
          ],
          "default": null,
          "description": "Maximum temperature (in Celsius)",
          "title": "Max Temperature"
        },
        "category": {
          "anyOf": [
            {
              "enum": [
                "Dry",
                "Fresh",
                "Frozen"
              ],
              "type": "string"
            },
            {
              "type": "null"
            }
          ],
          "default": null,
          "description": "Temperature control requirements",
          "title": "Category"
        }
      },
      "title": "TemperatureInfosData"
    },
    "TransportPriceData": {
      "type": "object",
      "properties": {
        "total_price": {
          "anyOf": [
            {
              "type": "number"
            },
            {
              "type": "null"
            }
          ],
          "default": null,
          "description": "Net price of the booking order (excluding taxes)",
          "title": "Total Price"
        },
        "currency": {
          "anyOf": [
            {
              "type": "string"
            },
            {
              "type": "null"
            }
          ],
          "default": null,
          "description": "Three-letter currency code (ISO 4217).",
          "title": "Currency"
        }
      },
      "title": "TransportPriceData"
    },
    "TruckData": {
      "type": "object",
      "properties": {
        "thinking": {
          "anyOf": [
            {
              "type": "string"
            },
            {
              "type": "null"
            }
          ],
          "default": null,
          "description": "Document the specific facts from the source that determine vehicle requirements. For example:\n- List any explicitly mentioned vehicle specifications\n- Note any temperature requirements stated\n- Record loading/unloading equipment mentioned\n- Include any access restrictions specified\n\nExample: 'Source specifies: refrigerated transport required, -20\u00b0C, loading dock not available, tail lift needed.'\n\nOnly include information that is explicitly stated in the source document. Do not make assumptions about requirements that aren't clearly specified.",
          "title": "Thinking"
        },
        "vehicle_type": {
          "anyOf": [
            {
              "enum": [
                "Tractor",
                "Carrier",
                "Light Vehicle"
              ],
              "type": "string"
            },
            {
              "type": "null"
            }
          ],
          "default": null,
          "description": "Primary vehicle classification:\n- 'Tractor': for semi-trailer trucks (Tracteur)\n- 'Carrier': for rigid trucks (Porteur) \n- 'Light Vehicle': for vans and smaller commercial vehicles (V\u00e9hicule l\u00e9ger)",
          "title": "Vehicle Type"
        },
        "body_type": {
          "anyOf": [
            {
              "enum": [
                "Tautliner",
                "Dry Van",
                "Refrigerated",
                "Flatbed",
                "Tanker"
              ],
              "type": "string"
            },
            {
              "type": "null"
            }
          ],
          "default": null,
          "description": "Type of cargo body/trailer:\n- 'Tautliner': curtainside trailer (Tautliner/Transpalette)\n- 'Dry Van': box trailer (Fourgon)\n- 'Refrigerated': temperature-controlled (R\u00e9frig\u00e9r\u00e9)\n- 'Flatbed': open platform (Plateau)\n- 'Tanker': liquid cargo (Citerne)",
          "title": "Body Type"
        },
        "tail_lift": {
          "anyOf": [
            {
              "type": "boolean"
            },
            {
              "type": "null"
            }
          ],
          "default": null,
          "description": "Indicates if the vehicle is equipped with a hydraulic tail lift for ground-level loading/unloading without dock access (Hayon \u00e9l\u00e9vateur)",
          "title": "Tail Lift"
        },
        "crane": {
          "anyOf": [
            {
              "type": "boolean"
            },
            {
              "type": "null"
            }
          ],
          "default": null,
          "description": "Indicates if the vehicle is equipped with a loading crane/HIAB for self-loading and unloading of heavy cargo (Grue)",
          "title": "Crane"
        }
      },
      "title": "TruckData"
    }
  }
}
```
</CodeGroup>



### 2 - Extract data from your documents with our Python SDK


To get started, install the `uiform` package using pip:

```bash
pip install uiform
```

Then, populate your `env` variables with your API keys:

<CodeGroup>
```python jupyter
import os
os.environ["OPENAI_API_KEY"] = "YOUR-API-KEY" # Your AI provider API key. Compatible with OpenAI, Anthropic, xAI.
os.environ["UIFORM_API_KEY"] = "sk_xxxxxxxxx"
```
```bash .env
OPENAI_API_KEY=YOUR-API-KEY # Your AI provider API key. Compatible with OpenAI, Anthropic, xAI.
UIFORM_API_KEY=sk_xxxxxxxxx
```
</CodeGroup>

Use the `UiForm` client to extract data from your documents:

<CodeGroup>
```python main.py {2, 9-15}
import json
from uiform.client import UiForm

with open("booking_confirmation_json_schema.json", "r") as f:
    json_schema = json.load(f)

document = "booking_confirmation.jpg"

client = UiForm()
response = client.documents.extract(
    json_schema = json_schema,
    document = document,
    model="gpt-4o-mini-2024-07-18",
    temperature=0
)
```
```json response.json
{'id': 'e94233ed-eb98-4ae0-861a-a9dfbbe28e56',
 'model': 'gpt-4o-mini-2024-07-18',
 'seed': None,
 'temperature': 0.0,
 'document': {'id': 'cf908729402d0796537bb91e63df5e339ce93b4cabdcac2f9a4f90592647e130',
  'name': 'acme_corp.jpg',
  'size': 1335913,
  'mime_type': 'image/jpeg'},
 'extraction': {'thinking': 'Analysis shows:\n- Clear booking confirmation document from ACME Corporation.\n- Contains 1 distinct shipment under single booking reference.\n- Client identified as ACME Corporation from letterhead.\n- Total price of 1500 EUR HT clearly stated.\n- All required signatures and references present.',
  'booking_id': 'SHIP-001',
  'payment': {'total_price': 1500.0, 'currency': 'EUR'},
  'client': {'company_name': 'ACME Corporation',
   'SIREN': None,
   'VAT_number_thinking': 'Found VAT number GB123456789 in header next to company name.',
   'VAT_number': 'GB123456789',
   'city': 'London',
   'postal_code': 'WC2N 5DU',
   'country': 'GB',
   'code': None,
   'email': 'client@acme.com'},
  'number_of_shipments': 1,
  'shipments_thinking': 'Shipment Analysis:\n1: 50 pallets, 5 units, total volume 1.44 m3, weight 500 kg, temperature range 5°C - 25°C (Fresh), dangerous goods UN1202 Class 3 - Liquid flammable.\n2: Pickup from ACME Corporation, 123 Elm Street, Suite 500, London, Greater London, WC2N 5DU, GB.\n3: Delivery to Beta Industries, 789 Pine Street, Munich, Bavaria, 80331 DE, +49 89 123456.',
  'shipments': [{'thinking': 'Source shows:\n- Single shipment identified by reference number SHIP-001.\n- Pickup from ACME Corporation to Beta Industries.\n- Must deliver between 10:00 and 16:00 on 5/3/2023.\n- References found: Client Ref. BC-67890.',
    'shipment_id': 'SHIP-001',
    'sender': {'company_name': 'ACME Corporation',
     'address': {'city': 'London',
      'postal_code': 'WC2N 5DU',
      'country': 'GB',
      'line1': '123 Elm Street',
      'line2': 'Suite 500'},
     'phone_number': '+44 20 7946 0958',
     'email_address': 'client@acme.com',
     'pickup_datetime': {'date': '2023-05-02',
      'start_time': '08:00:00',
      'end_time': '12:00:00'},
     'observations': 'Pickup at ACME Corporation, ensure all documents are ready.'},
    'recipient': {'company_name': 'Beta Industries',
     'address': {'city': 'Munich',
      'postal_code': '80331',
      'country': 'DE',
      'line1': '789 Pine Street',
      'line2': None},
     'phone_number': '+49 89 123456',
     'email_address': None,
     'delivery_datetime': {'date': '2023-05-03',
      'start_time': '10:00:00',
      'end_time': '16:00:00'},
     'observations': 'Delivery to Beta Industries, ensure unloading equipment is available.'},
    'goods': {'thinking': 'Source indicates:\n- 50 pallets of liquid flammable goods.\n- Reference: UN1202, Class 3.\n- Requires temperature control between 5°C and 25°C.\n- Ensure proper handling and safety measures are in place.',
     'packing': {'units': 50,
      'packing_type': 'pallet',
      'supplementary_parcels': None,
      'pallets_on_ground_thinking': 'Total of 50 pallets, each standard EUR size.',
      'pallets_on_ground': 50.0,
      'number_eur_pallet': 50,
      'observation': 'All pallets are standard EUR pallets.'},
     'dimensions': {'loading_meters_thinking': 'Using provided value of 1.44 m3 for volume calculation.',
      'loading_meters': None,
      'volume_thinking': 'Using provided value of 1.44 m3.',
      'volume': 1.44},
     'weight': 500.0,
     'temperature_infos': {'min_temperature': 5.0,
      'max_temperature': 25.0,
      'category': 'Fresh'},
     'dangerous_goods_infos': [{'thinking': 'Source indicates:\n- Dangerous goods UN1202, Class 3 - Liquid flammable.\n- Requires special handling and documentation.',
       'weight': None,
       'UN_code': 1202,
       'UN_label': 'Flammable Liquid',
       'UN_class': '3',
       'UN_packing_group': 'II',
       'ADR_tunnel_code': 'C'}]},
    'transport_constraints': {'thinking': 'Source specifies: refrigerated transport required, temperature control between 5°C and 25°C.',
     'vehicle_type': 'Carrier',
     'body_type': 'Refrigerated',
     'tail_lift': None,
     'crane': None}}]},
 'likelihoods': {'thinking': 0.9921241056702531,
  'booking_id': 0.9998840663508249,
  'payment': {'total_price': 0.9975170877384328,
   'currency': 0.9997090321193669},
  'client': {'company_name': 0.9999784691637917,
   'SIREN': 0.9995557247195026,
   'VAT_number_thinking': 1.0,
   'VAT_number': 0.9995085677923832,
   'city': 0.9839733329083605,
   'postal_code': 0.9838471985641967,
   'country': 0.3762832103888945,
   'code': 0.9995557247195026,
   'email': 0.9999932502087799},
  'number_of_shipments': 0.9999998063873687,
  'shipments_thinking': 0.9278115991168213,
  'shipments': [{'thinking': 0.8516071329998041,
    'shipment_id': 0.9998840663508249,
    'sender': {'company_name': 0.9999784691637917,
     'address': {'city': 0.9839733329083605,
      'postal_code': 0.9838471985641967,
      'country': 0.3762832103888945,
      'line1': 0.9993515157568538,
      'line2': 1.0},
     'phone_number': 0.9918943960091551,
     'email_address': 0.9999932502087799,
     'pickup_datetime': {'date': 0.999859520878116,
      'start_time': 0.9999589203757908,
      'end_time': 1.0},
     'observations': 0.3397073049247005},
    'recipient': {'company_name': 0.9999998063873687,
     'address': {'city': 0.9999895549275502,
      'postal_code': 0.9999998063873687,
      'country': 0.9999964686909351,
      'line1': 0.9999963494876631,
      'line2': 1.0},
     'phone_number': 0.9999934886141991,
     'email_address': 0.9995557247195026,
     'delivery_datetime': {'date': 0.9840801057114612,
      'start_time': 0.9999998063873687,
      'end_time': 0.9999994487765019},
     'observations': 0.2636216889390888},
    'goods': {'thinking': 1.0,
     'packing': {'units': 0.9955048834152134,
      'packing_type': 0.9986428313874298,
      'supplementary_parcels': 0.9995557247195026,
      'pallets_on_ground_thinking': 0.9999338889494318,
      'pallets_on_ground': 0.9955048834152134,
      'number_eur_pallet': 0.9955048834152134,
      'observation': 1.0},
     'dimensions': {'loading_meters_thinking': 0.2887916457038862,
      'loading_meters': 0.9995557247195026,
      'volume_thinking': 0.619517921101275,
      'volume': 0.9999882437011058},
     'weight': 0.9999604699583327,
     'temperature_infos': {'min_temperature': 0.9999946806438478,
      'max_temperature': 0.9999998063873687,
      'category': 0.9999897933310884},
     'dangerous_goods_infos': [{'thinking': 1.0,
       'weight': 0.9995557247195026,
       'UN_code': 0.9900534959502764,
       'UN_label': 0.9668220801782701,
       'UN_class': 0.994021364595207,
       'UN_packing_group': 0.9202036175532703,
       'ADR_tunnel_code': 0.44264760479416654}]},
    'transport_constraints': {'thinking': 1.0,
     'vehicle_type': 0.9690163928001971,
     'body_type': 0.6775101735710446,
     'tail_lift': 0.9995557247195026,
     'crane': 0.9995557247195026}}]},
 'regex_instruction_results': None,
 'additional_context_message': '',
 'usage': {'model': 'gpt-4o-mini-2024-07-18',
  'input_tokens': 13835,
  'output_tokens': 877,
  'cost': 0.00260145},
 'modality': 'native'}
 ```
</CodeGroup>
And that's it ! You can start processing documents at scale ! 
You have 1000 free requests to get started, and you can subscribe to the pro plan to get more.

But this minimalistic example is not much more useful than the bare openAI API. Continue reading to learn more about how to use UiForm **to its full potential**.

---

## Go further

- [Additional parameters](https://docs.uiform.com/document-api/additional-parameters)
- [Finetuning](https://docs.uiform.com/document-api/finetuning)
- Prompt optimization (coming soon)
- Data-Labelling with our AI-powered annotator (coming soon)

---

### Examples

You can view minimal notebooks that demonstrate how to use UiForm to process documents:

- [Quickstart](https://github.com/UiForm/uiform/blob/main/notebooks/Quickstart.ipynb)
- [Schema Creation](https://github.com/UiForm/uiform/blob/main/notebooks/Schema_creation.ipynb)
- [Finetuning](https://github.com/UiForm/uiform/blob/main/notebooks/Finetuning.ipynb)
