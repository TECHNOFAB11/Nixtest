# Reference

## `flakeModule`

The `flakeModule` for [flake-parts](https://flake.parts).

## `lib`

### `module`

The nix module for validation of inputs etc.
Used internally by `mkNixtestConfig`.

### `autodiscover`

```nix
autodiscover {
  dir,
  pattern ? ".*_test.nix",
}
```

Finds all test files in `dir` matching `pattern`.
Returns a list of modules (can be passed to `mkNixtest`'s `modules` arg).

### `mkNixtestConfig`

```nix
mkNixtestConfig {
  modules,
  args ? {},
}
```

Evaluates the test `modules`.
`args` are passed to the modules using `_module.args = args`.

**Noteworthy attributes**:

- `app`: nixtest wrapper
- `finalConfigJson`: derivation containing the tests json file

### `mkNixtest`

```nix
mkNixtest {
  modules,
  args ? {},
}
```

Creates the nixtest wrapper, using the tests in `modules`.
Basically `(mkNixtestConfig <arguments>).app`.
