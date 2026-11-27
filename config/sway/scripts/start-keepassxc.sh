#!/bin/sh
swaymsg -t subscribe -m '["window"]' |
  jq -n --unbuffered 'first(inputs | select(.change == "new" and .container.app_id == "org.keepassxc.KeePassXC")) | .container.id' |
  {
    keepassxc &
    read -r id && swaymsg "[con_id=$id] move container to workspace number 3"
  }
