# YAL: Yet Another Launcher

This is yet another CK3 launcher. The aim is to manage mods, easily import/export mods collections from/to Paradox Launcher or [Irony Mod Manager](https://github.com/bcssov/IronyModManager), and maybe in the future help warning about or fixing mod conflicts.

Why another launcher? Because I think the Paradox Launcher is quite limited and even dangerous to use. I've lost countless mod collections and it does play well with the steam workshop. The Irony Mod Manager is an amazing project, but the entry barrier is quite high, especially since I know neither C# nor Avalonia. This is mostly a solo project - mad props to bcssov for this AMAZING undertaking - and its aim is to manage many Paradox games, though the program has more features for Stellaris now in 2022. I wanted to do something a bit more simple, with a narrower scope.

## Scope of the project

This list is ordered by importance.

1. Manage mods [WIP]
    * import/export from/to Paradox Launcher
    * import/export from/to Irony Mod Manager
    * link mods to [Steam Workshop](https://steamcommunity.com/app/1158310/workshop) and/or [Paradox Mods](https://mods.paradoxplaza.com/games/ck3)
2. Detect mod conflicts [Planned]
    * Detect files that could override each other
3. Create a compatibility patch [Planned]
    * Offer a way to help in creating mod compatibility patches

## Technical implementation

This project uses [Python](https://python.org) and [PyQt](https://pypi.org/project/PyQt6)

Why? I think this will be a small/medium size project, with quite a lot of OS interaction on different platforms (Windows/MacOS/Linux). I thought this was the kind of project where Go would fit very well. Go is a language with a low-entry barrier, which would allow new programmers to contribute. The problem is that cross-platform UI frameworks are still quite young in Go. The one I tried, [Fyne](https://fyne.io), was nice and easy to start with but required me to implement too many low-level optimizations for image scaling, caching, etc.

I decided to use [Qt](https://doc.qt.io) instead. Qt is a well-know cross-platform C++ UI framework. C++ is a language with a high-level entry barrier, one I do not really enjoy using. You need to manage memory allocation yourself, and it has far too many bells and whistles and offer hundreds of different ways of achieving the smallest feat. Luckily there are Python bindings for Qt, and though I do not know Python, it has a very low-level entry barrier. Hopefull people will be easily able to join in or create plugins for this project!
