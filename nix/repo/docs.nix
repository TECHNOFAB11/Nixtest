{inputs, ...}: let
  inherit (inputs) pkgs doclib ntlib;

  optionsDoc = doclib.mkOptionDocs {
    module = ntlib.module;
    roots = [
      {
        url = "https://gitlab.com/TECHNOFAB/nixtest/-/blob/main/lib";
        path = "${inputs.self}/lib";
      }
    ];
  };
  optionsDocs = pkgs.runCommand "options-docs" {} ''
    mkdir -p $out
    ln -s ${optionsDoc} $out/options.md
  '';
in
  (doclib.mkDocs {
    docs."default" = {
      base = "${inputs.self}";
      path = "${inputs.self}/docs";
      material = {
        enable = true;
        colors = {
          primary = "green";
          accent = "light green";
        };
        umami = {
          enable = true;
          src = "https://analytics.tf/umami";
          siteId = "716d1869-9342-4b62-a770-e15d2d5c807d";
          domains = ["nixtest.projects.tf"];
        };
      };
      macros = {
        enable = true;
        includeDir = toString optionsDocs;
      };
      config = {
        site_name = "Nixtest";
        site_url = "https://nixtest.projects.tf";
        repo_name = "TECHNOFAB/nixtest";
        repo_url = "https://gitlab.com/TECHNOFAB/nixtest";
        extra_css = ["style.css"];
        theme = {
          logo = "images/logo.svg";
          icon.repo = "simple/gitlab";
          favicon = "images/logo.svg";
        };
        nav = [
          {"Introduction" = "index.md";}
          {"Usage" = "usage.md";}
          {"Reference" = "reference.md";}
          {"CLI" = "cli.md";}
          {"Example Configs" = "examples.md";}
          {"Options" = "options.md";}
        ];
        markdown_extensions = [
          "pymdownx.superfences"
          "admonition"
        ];
      };
    };
  }).packages
