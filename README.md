# ComposerGraph

This library allows you to build a dependency graph for an installed composer project.

## Author
This project is copy of https://github.com/schmittjoh/composer-deps-analyzer rewritten in golang

## Usage

Usage is quite simple:

```
a := composergraph.NewAnalyzer()
dgraph, err := a.Analyze(pathToPhpProject)
```

`dgraph` is a directed graph with the packages as nodes.

