Generate the static site from the manifest.yaml file and output it to the specified directory.

Usage:
  gengo generate [flags]

Flags:
  -h, --help                   help for generate
      --manifest stringArray   Path to the manifest file (default [gengo.yaml])
      --output string          Output directory (default "output")
      --plain                  Plain output. Useful for non-interactive shell
      --watch                  Enable watch mode with hot reload