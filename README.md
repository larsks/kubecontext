# kubecontext

Enable per-directory context settings for `kubectl` (or `oc`).

## File format

A `.kubecontext` file is a YAML file that may contain one or more of the following keys:

- `context` -- the name of a context in your `$KUBECONFIG` file.
- `namespace` -- the name of a namespace in your cluster
- `environment` -- a dictionary of environment variable names and
  values

For example:

```
context: my-cluster
namespace: project1
```

## Usage

Kubecontext looks for a file named `.kubecontext` in the current
directory and in all parents directories. It will then iterate over
any discovered `.kubecontex` files, applying them from the highest in
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