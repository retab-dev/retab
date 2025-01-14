# Cubeblocks

# Current spirit: 
We have two classes for each object, the class and its data surrogate: 
Classes are more pythonic, and the data is a pydantic class just for data validation purposes. 

# URGENT: 
Make the mail index match when you save then, otherwise you will loose them (test_io.py)


# Todo: 
Allow to select more finely where are going the attachments saved


# Todo: 
Write parsing connectors that allow to input/output data from pydantic classes to their related objects
- goods Done

# Todo: 
Write the pydantic class for the road class, for the trucks

# Todo: 
Write a more thorough suite of validation tests for dataprocess. You can use standard libraries like stdnum, phonenumbers, etc.. for that, or the csv files in the library

# Todo: Integrate the ISO 3166-2/1998 for the administrative subdivisions
