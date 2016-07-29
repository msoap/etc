#!/bin/sh

COWSAY_SELECT=$(cowsay -l | awk 'NR > 1' | sed 's/ /\n/g' | perl -pe 's/^(.+)$/<option value=$1>$1<\/>/')
FIGLET_SELECT=$(figlist | awk '$1 ~ /banner/, $1 ~ /term/' | perl -pe 's/^(.+)$/<option value=$1>$1<\/>/')

shell2http -port 8080 -form \
    / "echo '<html><h3>Cow&figlet</h3><form action=/cowsay>Cow: <input type=text name=msg><select name=kind>$COWSAY_SELECT</select><input type=submit></form> <form action=/figlet>Figlet: <input type=text name=msg><select name=kind>$FIGLET_SELECT</select><input type=submit></form>'" \
    /cowsay 'cowsay -f "$v_kind" "$v_msg"' \
    /figlet 'figlet -f "$v_kind" "$v_msg"' \
    /cowsay_list "cowsay -l | awk 'NR > 1' | sed 's/ /\n/g'" \
    /figlet_list "figlist | awk '\$1 ~ /banner/, \$1 ~ /term/'"
