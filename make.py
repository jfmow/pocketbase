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
#os.chdir('installer')
#subprocess.run(['go', 'build', '-o', 'installer', '.'])
#os.chdir('..')

# Create a zip file 'base.zip' containing 'base' and 'installer'
with zipfile.ZipFile('base.zip', 'w') as zipf:
    zipf.write('base')
    zipf.write(os.path.join('installer', 'installer'), 'installer')
#hostname = os.getenv("SSH_IP")
#username = os.getenv("SSH_USER")
#remote_path = os.getenv("SSH_PATH")
#private_key_path = os.getenv('SSH_DIR')
#passphrase = os.getenv('SSH_KEY')
#
#transport = paramiko.Transport((hostname, 698))
#private_key = paramiko.RSAKey(filename=private_key_path, password=passphrase)
#transport.connect(username=username, pkey=private_key)
#
#sftp = transport.open_sftp_client()
#sftp.put('base.zip', '/home/pie/pocketbase_docker_scripts/noti_db_priv/base.zip')
#
#sftp.close()
#
#transport.close()
#
#
#print("Build and packaging completed.")
