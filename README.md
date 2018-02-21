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


### Autocompletion

The releases page also contain autocomplete scripts for zsh and bash. These are redistributed from the [urfave/cli](https://github.com/urfave/cli) project.

To enable autocompletion source the file with the PROG environment variable set to prj.

```
PROG=prj source /path/to/bash_completion
```

#### Shell helper

If you want to use the `prj goto` command without having to type out the `eval $(prj goto projectname)` you can create the following function in your .bashrc/.zshrc

    function goto {
        eval $(prj goto $@)
    }
    
Then you can simply type `goto projectname`

This does not support the bash completion
