import os
import subprocess
import requests
import sys
from dotenv import load_dotenv
load_dotenv()
# Set environment variables
os.environ['GOOS'] = 'linux'
if len(sys.argv) == 1 or sys.argv[1] != "noarm":
    print('Using arm for build')
    os.environ['GOARCH'] = 'arm'



# Build the main.go in the current directory and name it 'base'
subprocess.run(['go', 'build', '-o', 'base', '.'])

def uploadToPocketBase(): 
    url = os.getenv('DATABASE_AUTH_URL')

    # Set the body parameters
    body_params = {
        'identity': os.getenv('DATABASE_DEV_USER'),
        'password': os.getenv('DATABASE_DEV_PASSWORD')
    }

    # Make the POST request with the body parameters
    response = requests.post(url, json=body_params)

    # Check the response
    if response.status_code == 200:
        # Assuming the response is in JSON format and contains a 'token' key
        response_json = response.json()
        if 'token' in response_json:
            print("Signed in!")
        else:
            print("Token not found in the response.")
    else:
        print("Error:", response.status_code)
        print(response.text)

    # Set the API endpoint where you want to post the file
    url = os.getenv("DATABASE_UPLOAD_URL")

    # Specify the file you want to upload
    files = {'base': ('base', open('base', 'rb'))}

    # Make the POST request with the files parameter
    response2 = requests.post(url, files=files, headers={'Authorization': response_json['token']})

    # Check the response
    if response2.status_code == 200:
        print("File uploaded successfully")
    else:
        print("Error uploading file. Status code:", response2.status_code)
        print(response.text)

continueInput2 = input("Do you wish to upload the base to the database? [y/n]\n")    
if continueInput2.lower() == "y":
    uploadToPocketBase()

print("Build and packaging completed.")