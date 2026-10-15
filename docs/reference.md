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

### `asserts` / `expect`

Assertion matchers for `type = "assertions"` tests (also available as `ntlib.helpers.expect` and `ntlib.helpers.asserts`):

- `eq expected actual` / `equal`: checks `actual == expected`.
- `ne expected actual` / `notEqual`: checks `actual != expected`.
- `gt actual min` / `greaterThan`: checks `actual > min`.
- `gte actual min` / `greaterThanOrEqual`: checks `actual >= min`.
- `lt actual max` / `lessThan`: checks `actual < max`.
- `lte actual max` / `lessThanOrEqual`: checks `actual <= max`.
- `contains container item`: checks if a list contains an element, or a string contains a substring.
- `notContains container item`: checks a list does not contain an element, or a string does not contain a substring.
- `hasAttr attr set`: checks if an attrset has the given attribute.
- `missingAttr attr set`: checks if an attrset lacks the given attribute.
- `hasAttrs attrs set`: checks if an attrset has all attributes in the given list.
- `matches regex str`: checks if a string matches the given regex pattern.
- `isNull val` / `isNotNull`: null checks.
- `isEmpty val` / `isNotEmpty`: checks if list, string, or attrset is empty / not empty.
- `throws expr` / `fails`: asserts that evaluating `expr` throws or aborts.
- `custom cond message` / `that`: custom condition with failure explanation message.
