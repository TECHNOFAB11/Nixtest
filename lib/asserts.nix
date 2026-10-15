{lib, ...}: let
  inherit (lib) generators;

  # Format any value to a readable string safely
  formatVal = val:
    if builtins.isString val
    then val
    else if builtins.isBool val || builtins.isInt val || builtins.isFloat val
    then builtins.toString val
    else generators.toPretty {} val;

  mkFailure = {
    message ? "Assertion failed",
    expected ? null,
    actual ? null,
  }: {
    ok = false;
    inherit message;
    expected =
      if expected != null
      then formatVal expected
      else "";
    actual =
      if actual != null
      then formatVal actual
      else "";
  };
in rec {
  inherit formatVal;

  equal = expected: actual:
    if actual == expected
    then true
    else
      mkFailure {
        message = "Expected values to be equal";
        inherit expected actual;
      };
  eq = equal;

  notEqual = expected: actual:
    if actual != expected
    then true
    else
      mkFailure {
        message = "Expected values to not be equal";
        expected = "not ${formatVal expected}";
        inherit actual;
      };
  ne = notEqual;

  greaterThan = actual: min:
    if actual > min
    then true
    else
      mkFailure {
        message = "Expected value to be greater than ${formatVal min}";
        expected = "> ${formatVal min}";
        inherit actual;
      };
  gt = greaterThan;

  greaterThanOrEqual = actual: min:
    if actual >= min
    then true
    else
      mkFailure {
        message = "Expected value to be greater than or equal to ${formatVal min}";
        expected = ">= ${formatVal min}";
        inherit actual;
      };
  gte = greaterThanOrEqual;

  lessThan = actual: max:
    if actual < max
    then true
    else
      mkFailure {
        message = "Expected value to be less than ${formatVal max}";
        expected = "< ${formatVal max}";
        inherit actual;
      };
  lt = lessThan;

  lessThanOrEqual = actual: max:
    if actual <= max
    then true
    else
      mkFailure {
        message = "Expected value to be less than or equal to ${formatVal max}";
        expected = "<= ${formatVal max}";
        inherit actual;
      };
  lte = lessThanOrEqual;

  contains = container: item:
    if builtins.isList container
    then
      if builtins.elem item container
      then true
      else
        mkFailure {
          message = "Expected list to contain item";
          expected = "to contain: ${formatVal item}";
          actual = container;
        }
    else if builtins.isString container
    then
      if lib.strings.hasInfix item container
      then true
      else
        mkFailure {
          message = "Expected string to contain substring";
          expected = "to contain: ${formatVal item}";
          actual = container;
        }
    else
      mkFailure {
        message = "Expected list or string for 'contains'";
        actual = container;
      };

  notContains = container: item:
    if builtins.isList container
    then
      if !(builtins.elem item container)
      then true
      else
        mkFailure {
          message = "Expected list to not contain item";
          expected = "not to contain: ${formatVal item}";
          actual = container;
        }
    else if builtins.isString container
    then
      if !(lib.strings.hasInfix item container)
      then true
      else
        mkFailure {
          message = "Expected string to not contain substring";
          expected = "not to contain: ${formatVal item}";
          actual = container;
        }
    else
      mkFailure {
        message = "Expected list or string for 'notContains'";
        actual = container;
      };

  hasAttr = attr: set:
    if !(builtins.isAttrs set)
    then
      mkFailure {
        message = "Expected attrset for 'hasAttr'";
        actual = set;
      }
    else if builtins.hasAttr attr set
    then true
    else
      mkFailure {
        message = "Expected attrset to have attribute '${attr}'";
        expected = "to have attribute: ${attr}";
        actual = "available attributes: ${formatVal (builtins.attrNames set)}";
      };

  missingAttr = attr: set:
    if !(builtins.isAttrs set)
    then
      mkFailure {
        message = "Expected attrset for 'missingAttr'";
        actual = set;
      }
    else if !(builtins.hasAttr attr set)
    then true
    else
      mkFailure {
        message = "Expected attrset to not have attribute '${attr}'";
        expected = "not to have attribute: ${attr}";
        actual = "found attribute '${attr}' with value: ${formatVal set.${attr}}";
      };

  hasAttrs = attrs: set:
    if !(builtins.isAttrs set)
    then
      mkFailure {
        message = "Expected attrset for 'hasAttrs'";
        actual = set;
      }
    else let
      missing = builtins.filter (a: !builtins.hasAttr a set) attrs;
    in
      if missing == []
      then true
      else
        mkFailure {
          message = "Expected attrset to have attributes ${formatVal attrs}";
          expected = "missing attributes: ${formatVal missing}";
          actual = "available attributes: ${formatVal (builtins.attrNames set)}";
        };

  matches = regex: str:
    if !(builtins.isString str)
    then
      mkFailure {
        message = "Expected string for 'matches'";
        actual = str;
      }
    else if builtins.match regex str != null
    then true
    else
      mkFailure {
        message = "Expected string to match regex '${regex}'";
        expected = "match regex: ${regex}";
        actual = str;
      };

  isNull = val:
    if val == null
    then true
    else
      mkFailure {
        message = "Expected value to be null";
        expected = "null";
        actual = val;
      };

  isNotNull = val:
    if val != null
    then true
    else
      mkFailure {
        message = "Expected value to not be null";
        expected = "not null";
        actual = "null";
      };

  isEmpty = val: let
    empty =
      if builtins.isList val
      then val == []
      else if builtins.isAttrs val
      then val == {}
      else if builtins.isString val
      then val == ""
      else false;
  in
    if empty
    then true
    else
      mkFailure {
        message = "Expected value to be empty";
        expected = "empty";
        actual = val;
      };

  isNotEmpty = val: let
    notEmpty =
      if builtins.isList val
      then val != []
      else if builtins.isAttrs val
      then val != {}
      else if builtins.isString val
      then val != ""
      else true;
  in
    if notEmpty
    then true
    else
      mkFailure {
        message = "Expected value to not be empty";
        expected = "not empty";
        actual = val;
      };

  throws = expr: let
    res = builtins.tryEval expr;
  in
    if !res.success
    then true
    else
      mkFailure {
        message = "Expected expression to throw an error, but it succeeded";
        expected = "to throw";
        actual = "evaluated to: ${formatVal res.value}";
      };
  fails = throws;

  custom = cond: message:
    if cond
    then true
    else
      mkFailure {
        inherit message;
      };
  that = custom;
}
