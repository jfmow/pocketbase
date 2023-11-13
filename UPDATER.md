# How to use the updater

Once you have a NEW version of the program you can upload it to pocketbases table in the db and run the following cmd with all envs set correctly.

## Command
Use  
```bash
 ./pb/base updateme
 ```
 or your equv of ./pb/base

## Enviroment var
You **MUST** set *updateApiUrl* & *updateApiKey*

updateApiUrl is: yourappurl.app/update/latest?auth=

Must have ?auth= as querry param

updateApiKey is a key to check access, like a password, it must be set in your env.

Example
- updateApiUrl="https://proti.suddsy.dev/update/latest?auth="
- updateApiKey=31f559830f6713c328a7b8bff94bab9024d5735b218b5bc9

# Warnings
The app auto reboots the system after downloading the update to make sure that no ports are occupied by the old app so recomend to run in docker enviroment