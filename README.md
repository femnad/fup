# fup

A [pyinfra](https://pyinfra.com/) based workstation initializer.

## Why?

Because that moment when you first start using a newly provisioned OS is like trying to walk without moving your legs.

For a better explanation and implementation see [comtrya](https://github.com/comtrya/comtrya).

## Why Not Ansible?

Because it causes me create monstrosities like [this](https://github.com/femnad/casastrap). Seriously, who writes hundreds of playbooks which do one single thing?

## Why Not SaltStack?

SaltStack is great and the enabler of the provisioning method I was most happy with: [anr](https://github.com/femnad/anr) which relies on `salt-ssh`. Since it's pretty flexible and expressive you can make it work as provisioner for your user's home and perform elevated privilege steps with dedicated states.

However, this is a finicky setup and needs some hacks, like resetting permissions on `/var/tmp` between runs ([embarrassing script](https://gitlab.com/femnad/chezmoi/-/blob/9c379c8105456d53bcf38de8410fc7193dafadce/bin/executable_salt-pre-flight)), always specifying the user or changing permissions for states because some states have to mix root and non-root operations. Also, when I last tried to provision with SaltStack I had hit two show-stopper bugs, one dependency pinning issue and another Python 3.10 incompatibility, meaning I was stuck with a non-usable OS state.

## So Another Python base provisioning tool is the answer then?

Well, not really, but it's what I can hack together in a relatively short time.

## No, Really, Why Not Comtrya Then?

I don't know, scratching an itch maybe?. Let me have this?
