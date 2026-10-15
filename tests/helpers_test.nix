{ntlib, ...}: {
  suites."Helpers Tests" = {
    pos = __curPos;
    tests = [
      {
        name = "assertion helpers";
        type = "assertions";
        assertions = with ntlib; {
          "equality" = asserts.eq 42 42;
          "not equal" = asserts.ne 42 100;
          "greater than" = asserts.gt 10 5;
          "greater than or equal" = asserts.gte 10 10;
          "less than" = asserts.lt 5 10;
          "less than or equal" = asserts.lte 5 5;
          "list contains" = asserts.contains [1 2 3] 2;
          "list not contains" = asserts.notContains [1 2 3] 4;
          "string contains" = asserts.contains "hello world" "world";
          "hasAttr" = asserts.hasAttr "a" {a = 1;};
          "missingAttr" = asserts.missingAttr "b" {a = 1;};
          "hasAttrs" = asserts.hasAttrs ["a" "b"] {
            a = 1;
            b = 2;
          };
          "matches regex" = asserts.matches "^v[0-9]+$" "v2";
          "isNull" = asserts.isNull null;
          "isNotNull" = asserts.isNotNull 1;
          "isEmpty" = asserts.isEmpty [];
          "isNotEmpty" = asserts.isNotEmpty [1];
          "throws" = asserts.throws (throw "oops");
          "raw boolean" = true;
          "expect alias" = expect.eq "foo" "foo";
        };
      }
    ];
  };
}
