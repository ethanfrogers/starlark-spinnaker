# starlark-spinnaker

## About

`starlark-spinnaker` is an example project to demonstrate how you might use Starlark
to render Spinnaker pipelines using code instead of YAML or JSON templating.

When using this project, one must implement a `main()` method which returns a `dict` (or map)
representing your Spinnaker pipeline.

If you aren't familiar with Starlark, check out some of these resources
* //TODO - add resources


## TODO
- [ ] Figure out a more use friendly way of handling stage dependencies instead of users specifying refIds manually
- [ ] Build `spinnaker.sky` directly into the tool as a predefined module
- [ ] Add more stage types i.e Deploy/Bake Manifest
- [ ] Add artifact support
- [ ] Make renderer into standalone CLI / refactor to be pretty


## Example Usage

`starlark-spinnaker` can be paired with the `spin` CLI to render and save pipelines. For example,
given a Starlark file like this

```build
def main():
    return {
        'application': 'myapp',
        'name': 'Deploy to Prod',
        'stages': [
            {
                'type': 'wait',
                'waitTime': 30,
                'refId': 0,
            }
        ]
    }
```

You can render the above example into a Spinnaker pipeline and save it like so:

```shell script
$ starlark-spinnaker --config /path/to/file.sky | spin pipeline save
```

You can now see your pipeline saved into Spinnaker.