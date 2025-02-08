# Notes
            
[//]: # (Small scale TODOs listed here. Anything bigger should be a GitHub Issue.)
                      
[//]: # (it's like - on 1st pass it won't but when certain unknown changes occur while
typing it will suddenly render the string. maybe the ordering of things needs to
change.) 

[//]: # (TODO: Finish out porting things from Zed Zui syntax)
[//]: # (TODO: Zed: Add string interpolation)
[//]: # (TODO: Zed: Add record parsing)
[//]: # (TODO: Try the original TextMate JSON with ZSON)
[//]: # (TODO: Folding by parens isn't working? - foldingStartMarker is ignored looks like )
[//]: # (TODO: Check GitHub issues)

## Build Process

Currently: edit json in-place, go through the manual Settings "reload" process:
`Open Settings -> Remove -> Apply -> Re-add -> Close.` (See Reloading in
RubyMine section below for more).

## References

[Zed Syntax in Zui source
code](https://github.com/brimdata/zui/blob/edbe753b548b56d140802aae65ae14a190ea5e42/apps/zui/src/core/zed-syntax.ts#L26).

1.5 TextMate [Example
Grammar](https://macromates.com/manual/en/language_grammars#example_grammar)

[Older Sublime
docs](https://sublime-text-unofficial-documentation.readthedocs.io/en/sublime-text-2/reference/syntaxdefs.html)
detailing Compatibility with TextMate language files, and is a fairly compact
reference.

This [community
thread](https://intellij-support.jetbrains.com/hc/en-us/community/posts/12784192098066/comments/12899228677010)
has a good pointer to the JetBrains implementation of the TextMate Bundles - but
it's out of date - [I found it in the latest
repo](https://github.com/JetBrains/intellij-community/blob/master/plugins/textmate/core/src/org/jetbrains/plugins/textmate/language/syntax/selector/TextMateSelectorParser.kt).
                                   
[This
class](https://github.com/JetBrains/intellij-community/blob/master/plugins/textmate/core/src/org/jetbrains/plugins/textmate/Constants.kt)
appears to show the keywords in the TextMate Bundles plugin. Which seems to
confirm that `file_extensions` isn't going to be read out of the .plist, but
fileTypes will.

[Sublime Syntax](https://www.sublimetext.com/docs/syntax.html) - this is clear
that this is their own way of doing it, and we shouldn't expect RubyMine to work
with these at all, which is sort of what I've seen so far. The definitions are
very different.

## Reloading in RubyMine

As I make changes, do I have to remove/re-add to force RubyMine to reload the
bundle? What's the minimum effort as I iterate on things.

Meh, looks like you have to remove and re-add in the Settings dialog. :/ Even
closing just the current project isn't enough. Presumably closing the whole IDE
would work, but that won't be fast.

```
WORKS: Open Settings -> Remove -> Apply -> Re-add -> Close.
NOPE:  Open Settings -> Remove -> Re-add -> Close.
```

If a file is visible behind the dialog, highlighting won't update until after
the Settings dialog is closed, not Apply.

## Notes on Names
                                             
see [Naming
Conventions](https://macromates.com/manual/en/language_grammars#naming_conventions)

I don't see much use of `support.*` - Java and C# in the included TextMate
Bundles plugin (sources) don't use it at all.

`entity.name.function` - appears to be for parsed function names (C#, Javascript)


