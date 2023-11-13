import os
import subprocess
import shutil
import zipfile
import paramiko
from dotenv import load_dotenv
load_dotenv()
# Set environment variables
os.environ['GOOS'] = 'linux'
os.environ['GOARCH'] = 'arm'

# Build the main.go in the current directory and name it 'base'
subprocess.run(['go', 'build', '-o', 'base', '.'])

# Create a subdirectory 'installer' and build main.go inside it, naming it 'installer'
#os.chdir('installera')
#subprocess.run(['go', 'build', '-o', 'installer', '.'])
#os.chdir('..')

# Create a zip file 'base.zip' containing 'base' and 'installer'
#with zipfile.ZipFile('base.zip', 'w') as zipf:
 #   zipf.write('base')
    #zipf.write(os.path.join('installera', 'installer'), 'installer')
continueInput = input("Do you wish to upload the base to noti_db_priv? [y/n]\n")    
if continueInput.lower() == "y":
    hostname = os.getenv("SSH_IP")
    username = os.getenv("SSH_USER")
    remote_path = os.getenv("SSH_PATH")
    private_key_path = os.getenv('SSH_DIR')
    passphrase = os.getenv('SSH_KEY')

    transport = paramiko.Transport((hostname, 698))
    private_key = paramiko.RSAKey(filename=private_key_path, password=passphrase)
    transport.connect(username=username, pkey=private_key)

    sftp = transport.open_sftp_client()
    sftp.put('base', '/home/pie/pocketbase_docker_scripts/noti_db_priv/base')

    sftp.close()

    transport.close()

print("Build and packaging completed.")
