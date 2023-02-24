# Python's Regular Expressions

<!--
In this example, we demonstrate how to document some code and use mlp to test the code shown in the documentation.

This block code is hidden in the human formatted document but will be inserted in the generated python file:

~~~py >
<<<import>>>
~~~
-->

Python has a builtin regular expression module called `re`. It supports many features of the perl regular expressions.

## Search

The most commonly used method in that module is `search()`. It takes the regular expression as first parameter and the
text where the regular expression is applied as second parameter. It returns `None` when the search fails and a `Match`
object when it succeeds.

~~~py >
match = re.search(r'^`{3,}\s?([\w\+]+)\s+>\s*([\w\.\-\/]*)$', '``` csv > data.csv')
~~~

<!--
check that the result is as expected, but don't show that to the end reader

~~~py >
assert match is not None
~~~
-->

The match groups may be retrieved with the `group()` method of the match object.

~~~py >
lang = match.group(1)
file = match.group(2)
~~~

<!--
~~~py >
assert lang == "csv"
assert file == "data.csv"
~~~
-->

## Import

Before being able to use all those great feature, we must import the `re` module.

~~~py "import"
import re
~~~

<!--
just run these commands to check the code in this file:

~~~sh
mlp python-example.md
python python-example.md.py
~~~

note that, because it doesn't have a `>`, the preceding block code doesn't generate any file.
-->

Enjoy!
