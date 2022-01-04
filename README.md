<p align="center">
    <img src="./.github/logo.png" width="100" alt="logo"/>
    <h1 align="center">JT</h1>
    <p align="center">a java tool for the command line</p>
    <p align="center">
        <a href="https://github.com/tsatke/jt/actions/workflows/go.yml"><img src="https://github.com/tsatke/jt/actions/workflows/go.yml/badge.svg"></a>
    </p>
</p>

---

`jt` is a java tool for the command line.
It is meant to be the JDT in the terminal (I know that it still has to go a long way).

## Install

You can either build it from source with `go build`, or install it with
```bash
go install github.com/tsatke/jt@latest
```

## Usage

After installing, the `jt` command should work.
Everything in `jt` is done with class paths, rather than qualified names (`java/lang/Object` rather than `java.lang.Object`).
In the future, this tool might support the qualified names as input, but for now, it just saves a lot of headaches if we just use the paths.

#### Verbose output

If you want more output (or think something might be wrong), you can check the debug output by adding the `-v` flag.

### Listing classes in a jar file

Assuming you have a jar file with 5 classes in it (no matter where exactly, just inside the jar file), the following works.
```bash
$ jt classes myjar.jar
com/mypackage/Class1
com/mypackage/Class2
com/mypackage/Class3
com/mypackage/Class4
com/mypackage/Class5
```

### Supported project formats

At the moment, `jt` supports Maven and Eclipse project formats.
That is, for Maven, it uses the `pom.xml` and for Eclipse the `.classpath` file to get a classpath.
On that classpath, `jt` will search for classes.
Currently, it will always consider the `JAVA_HOME` variable and use that as the standard library on any classpath, regardless what the Maven or Eclipse project have configured.
If `JAVA_HOME` is not set, you will not be able to get information about classes that are located in the standard library.

### Viewing the classpath

`jt` can display the classpath of a project.
The standard library is always on top of other entries on the classpath.
```bash
$ jt classpath
/path/to/my/project/src/main/java
/path/to/java-home/lib/rt.jar
...
/path/to/java-home/lib/charsets.jar
/path/to/maven-repo/junit/junit/4.11/junit-4.11.jar
/path/to/maven-repo/org/hamcrest/hamcrest-core/1.3/hamcrest-core-1.3.jar
```

### Finding classes

You can use `jt` to find classes on the classpath of the project that you are currently in.
```bash
$ jt find App
Project results:
com/mypackage/App
com/mypackage/App$Builder
Classpath results:
com/thirdparty/AppConfiguration
```
but
```bash
$ jt find App | cat
com/mypackage/App
com/mypackage/App$Builder
com/thirdparty/AppConfiguration
```
if you're not on a terminal, `jt` will not print the headers, so you can use `grep`, `xargs` and your other favorite tools as you're used to.

In addition, if you know that a specific class is in your project, and you don't need to search a (potentially) large classpath, you can pass the `--no-classpath` option.
This will keep `jt` from even building a classpath, and save you a lot of time, especially in Maven projects.

### Viewing superclasses

You can view the superclasses of a given class on the classpath.
Interfaces are not implemented yet.
```bash
$ jt superclass com/mypackage/io/SpecialInputStream
com/mypackage/io/SpecialInputStream
com/mypackage/io/AbstractInputStream
java/io/InputStream
java/lang/Object
```

### Viewing subclasses
not implemented yet

<div>Icons made by <a href="https://www.freepik.com" title="Freepik">Freepik</a> from <a href="https://www.flaticon.com/" title="Flaticon">www.flaticon.com</a></div>