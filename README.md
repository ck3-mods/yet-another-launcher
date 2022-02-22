# YAL: Yet Another Launcher

This is yet another CK3 launcher. The aim is to manage mods, easily import/export mods collections from/to Paradox Launcher or [Irony Mod Manager](https://github.com/bcssov/IronyModManager), and maybe in the future help warning about or fixing mod conflicts.

Why another launcher? Because I think the Paradox Launcher is quite limited and even dangerous to use. I've lost countless mod collections and it does play well with the steam workshop. The Irony Mod Manager is an amazing project, but the entry barrier is quite high, especially since I know neither C# nor Avalonia. This is mostly a solo project - mad props to bcssov for this AMAZING undertaking - and its aim is to manage many Paradox games, though the program has more features for Stellaris now in 2022. I wanted to do something a bit more simple, with a narrower scope.

## Scope of the project

This list is ordered by importance.

1. Manage mods [WIP]
    * import/export from/to Paradox Launcher
    * import/export from/to Irony Mod Manager
    * link mods to [Steam Workshop](https://steamcommunity.com/app/1158310/workshop/) and/or [Paradox Mods](https://mods.paradoxplaza.com/games/ck3)
2. Detect mod conflicts [Planned]
    * Detect files that could override each other
3. Create a compatibility patch [Planned]
    * Offer a way to help in creating mod compatibility patches

## Technical implementation

This project uses [Golang](https://go.dev/) and [Fyne UI toolkit](https://fyne.io).

Why? I though this would be a small/medium size project, with quite a lot of OS interaction on different platforms (Windows/MacOS/Linux). I think this is the kind of project where Go would fit very well. Go is a language with a low-entry barrier, which would allow new programmers to contribute. I think it is far easier to scale a project with Go than with Python because of the strict typing and the explicit error checking.

This is definitely not a personal inclination as I am not a fan of Go, I would rather learn and use Rust than Go. But I want this project to be as open as possible, and that requires a language that won't push people away. I also welcome the idea of learning another language, and Go seems like a good language to prototype or script something quickly.
