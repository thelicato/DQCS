<h1 align="center">
    <img src="./docs/logo.png?raw=true" width="400">
</h1>

<h4 align="center">Dummy QEMU Clipboard Sharing</h4>

<p align="center">
  <a href="#-features">Features</a> •
  <a href="#-usage">Usage</a> •
  <a href="#-installation">Installation</a> •
  <a href="#-license">License</a> •
</p>

---

``dqcs`` is a handy little tool for those who **love** QEMU but find themselves **cursing** the SPICE (performance issues). After stumbling upon SDL, I was amazed. No glitches, smooth as butter, nearly perfect. The only catch? No clipboard sharing... yet. So, I whipped up this quick-and-dirty fix using ``virtio-serial`` to fill that gap until SDL adds native clipboard support. Enjoy! Or don't. But I think you should at least try.


## ⚡ Features

The structure of the project is pretty simple:
- a component that runs on the host
- a component that runs on the guest
- a ``virtio-serial`` device for the communication between the two components.

The following is a list of the features (both planned and implemented):
- [x] Allow bidirectional copy between Linux host and Linux guest
- [x] Allow bidirectional copy between Linux host and Windows guest 
- [ ] Allow image clipboard sharing
- [ ] Add GUI with systray to show logs and status

## 📚 Usage

```
dqcs -h
```

This will display the help for the tool

```
        ░█▀▄░▄▀▄░█▀▀░█▀▀
        ░█░█░█\█░█░░░▀▀█
        ░▀▀░░░▀\░▀▀▀░▀▀▀

v0.1.0 - https://github.com/thelicato/dqcs

Dummy QEMU Clipboard Sharing for those who don't like SPICE

Usage:
  dqcs [flags]
  dqcs [command]

Available Commands:
  guest       Run the DQCS host component
  help        Help about any command
  host        Run the DQCS host component

Flags:
  -h, --help   help for dqcs

Use "dqcs [command] --help" for more information about a command.

```

The help is pretty self-explainatory, the correct order is the following:
1. Run ``dqcs host``
2. Start QEMU with a ``virtio-serial`` device
3. Run ``dqcs guest`` in the QEMU VM

### Configure QEMU

The following options are needed in your QEMU start command to make things work:

```
  -device virtio-serial \
  -chardev socket,path=<CHANGEME>,id=clipboard \
  -device virtserialport,chardev=clipboard,name=com.dqcs.clipboard
```

The ``path`` value needs to be set accordingly with the path specified when running the ``dqcs host`` command. 
Currently the name of the ``virtserialport`` device is hardcoded to ``com.dqcs.clipboard`` (could be customizable in the future).


### Linux Service Setup

Here is a sample Linux service that you can use to automate the usage:

```
[Unit]
Description=dqcs
After=graphical-session.target
Wants=graphical-session.target

[Service]
Type=simple
ExecStart=/usr/local/bin/dqcs guest
Restart=on-failure
RestartSec=10
User=root
Environment=PATH=/usr/bin:/bin
Environment=DISPLAY=:1
Environment=XAUTHORITY=/run/user/1000/gdm/Xauthority
WorkingDirectory=/root

[Install]
WantedBy=default.target
```

To create a Linux service:
1. Create a ``dqcs.service`` file in ``/etc/systemd/system`` with the above content
2. Reload the available service with ``systemctl daemon-reload``
3. Start the ``dqcs.service`` with ``systemctl start dqcs``
4. Check the current status of the service with ``systemctl status dqcs``
5. Enable autostart with ``systemctl enable dqcs``

### Windows Setup

Even though the tool is built to support Windows services the clipboard **does not seem to be updated** when running as a Service. For this reason I'm just running it as a *Scheduled Task*.

Is it ugly? Definitely.
Does it work? Hell yeah.

A solution could be to either fix the *Windows Service* or build a light GUI and keep it in the systray.


## 🚀 Installation

Run the following command to install the latest version:

```
go install github.com/thelicato/dqcs@latest
```

Or you can simply grab an executable from the [Releases](https://github.com/thelicato/dqcs/releases) page.

## 🪪 License

_dqcs_ is made with 🖤 and released under the [MIT LICENSE](./LICENSE).