import os
import subprocess
import zipfile
import sys

def build_and_zip(target_arch=''):
    # Set environment variables for GOOS and GOARCH
    if target_arch:
        os.environ['GOARCH'] = target_arch
    os.environ['GOOS'] = 'linux'

    # Build the main.go in the current directory and name it 'base'
    build_command = "go build -o base ."
    subprocess.run(build_command, shell=True, check=True)

    # Build the main.go in the 'installer' directory and name it 'installer'
    os.chdir('installer')
    build_command = "go build -o installer ."
    subprocess.run(build_command, shell=True, check=True)
    os.chdir('..')

    # Create a zip file called 'base.zip' and add 'base' and 'installer' to it
    with zipfile.ZipFile('base.zip', 'w') as zipf:
        zipf.write('base', arcname='base')
        zipf.write(os.path.join('installer', 'installer'), arcname='installer')
        zipf.write('preview_page.json', arcname='preview_page.json')
    os.remove(os.path.join('installer', 'installer'))

def git_commit_push_sync():
    # Check if 'base.zip' exists
    if not os.path.exists('base.zip'):
        print("Error: 'base.zip' not found.")
        return

    # Add, commit, and push only the 'base.zip' file
    subprocess.run(["git", "add", "base.zip"], check=True)
    subprocess.run(["git", "commit", "-m", "Update base.zip"], check=True)
    subprocess.run(["git", "push"], check=True)

if __name__ == "__main__":
    target_architecture = sys.argv[1] if len(sys.argv) > 1 else None
    build_and_zip(target_architecture)
    git_commit_push_sync()
