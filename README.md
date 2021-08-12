# kubecontext

Enable per-directory context settings for `kubectl`.

## Overview

Install the `kubecontext` binary as `$HOME/bin/k`, create a
`.kubecontext` file in your project directory, like this...

```
context: my-cluster
namespace: my-project
```

...and then run `k` instead of `kubectl`:

```
$ k get pod
```

Etc.

## File format

A `.kubecontext` file is a YAML file that may contain one or more of the following keys:

- `context` -- the name of a context in your `$KUBECONFIG` file. You
  can get a list of available contexts by running `kubectl config
  get-contexts`.
- `namespace` -- the name of a namespace in your cluster
- `command` -- name of a command to run instead of `kubectl` (e.g.,
  `oc`)
- `environment` -- a dictionary of environment variable names and
  values. I use this primarily for setting proxy information (because
  `oc login` delete proxy settings from the kubeconfig file).

For example:

```
context: my-cluster
namespace: project1
```

## Usage

Kubecontext looks for a file named `.kubecontext` in the current
directory and in all parents directories. It will then iterate over
any discovered `.kubecontext` files, applying them from the highest in
the hierarchy to the lowest.

For example, assume the following file hierarchy:

```
/home/you/projects/project1
  .kubecontext
  app1/
    .kubecontext
  app2/
    .kubecontext
```

`/home/you/projects/project1/.kubecontext` looks like this:

```
context: my-cluster
```


`/home/you/projects/project1/app1/.kubecontext` looks like this:

```
namespace: app1
```

`/home/you/projects/project1/app2/.kubecontext` looks like this:

```
namespace: app2
```

If you were to run `kubecontext get pod` in `/home/you/project1/app1`,
that would result in the following sequence of commands:

```
kubectl config use-context my-cluster
kubectl config set-context --current --namespace=app1
kubectl get pod
```

If you were to `cd ..` into the `project1` directory and run
`kubecontext get pod`, that would run:

```
kubectl config use-context my-cluster
kubectl get pod
```

In other words, it wouldn't change your namespace because the
`.kubecontext` file in that directory does not set a namespace.
