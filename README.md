# Prj

Prj is a command line tool for managing project directories

A linux x86_64 binary can be downloaded from the releases page. Other platforms can build from the source.

Just running the prj command will output the help text that should get you started.

## Config

Permanent configuration is stored under $HOME/.prj/db.json

To see available config options run

    prj config list
    
So to set AlwaysGit to true (create git repository for every project by default)

    prj config set AlwaysGit true

