## grofi

Small rofi script to search for Go packages on `pkg.go.dev`.

Usage:
```
git clone https://github.com/giulianopz/grofi.git ~/.grofi
cd ~/.grofi
go build -o /usr/local/bin/grofi .
chmod u+x /usr/local/bin/grofi
```

Then add a custom keyboard shortcut to launch `rofi` in script mode according to your desktop environment (e.g. see [here](https://docs.fedoraproject.org/en-US/quick-docs/proc_setting-key-shortcut/) for GNOME):
```
rofi -show grofi -modi "grofi:/usr/local/bin/grofi" 
``` 

### References:
- [rofi-script(5)](https://man.archlinux.org/man/rofi-script.5.en)
- [Rofi based scripts](https://github.com/davatorium/rofi-scripts)
