# GoObfuscator

My thinking is using the built-in go AST tools so that I can 
eventually do more than simple search/replace type things, but 
instead inject different techniques based on what the "node" is in the AST.

Did we find a function? ok, what function specific things can we do to complicate things? A string var? let's inject some custom encoder/decoder to slow things down a bit. 
This is just for fun and a reason to get back to hacking on go projects, but could be fun.

Currently the only thing that works (somewhat) is renaming structs (and their properties) with random gibberish.
