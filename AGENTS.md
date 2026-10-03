# AGENTS.md

## Project overview

Go library. The public API lives in subpackages;
the `internal` package is not public.
the root package holds only the module overview.

## Constraints
- Treat exported identifiers as a stable contract and avoid breaking changes.
- Do not use third-party dependencies except for testing purposes.
- Panic only to indicate programmer errors like using uninitialized values, etc.
- Never panic on runtime failures; return errors instead.

## Build and test commands

```bash
# Install development tools
make tools

# Install dependencies
make tidy

# Check dependencies
make tidy-check

# Format code
make fmt

# Check code formatting
make fmt-check

# Check code correctness
make vet

# Run the linter
make lint

# Run all tests with race detector
make test

# Check code coverage
make cover

# Check dependencies, code formatting, and code correctness;
# Run the linter and all tests with race detector.
# This command must be executed after every change.
make check

# Clean the project
make clean
```

## Code style guidelines
- Follow standard Go conventions.
- Adopt the style of the existing code.
- Format code with `make fmt`.
- Exported identifiers in the public API must have doc comments starting with the identifier name.

## Testing instructions
- Tests must cover the public API.
- Tests must pass `make test`; run `make test` multiple times to ensure stability.
- Add or update tests for any code change.
- Keep tests in `_test.go` files alongside the code, in the external test package,
  so they use only the public API.
- Adopt the style of the existing tests.
- Organize tests as named subtests using the `t.Run` function.
- Assert test results with `require` and `assert` from the `testify` package.
- Prefer `require` over `assert`; use `assert` only in go-routines started within the same test.
- Usage examples must be provided for the public API as `Example` functions.
- Keep `Example` functions in `*_example_test.go` files.
- Verify the result of an `Example` function with an `// Output:` comment.
- Check coverage with `make cover`; it must not decrease.

## Definition of done
- Code must pass `make check` with no errors or warnings.
- New exported top-level identifiers must be briefly mentioned in `doc.go`.
- New exported top-level identifiers must be listed in `README.md`.
- Usage examples must be provided in the corresponding sections of `README.md`.
- Usage examples in `README.md` must compile.

## Repository and branches
- Work in the current branch; do not create or switch branches.
- Never commit, tag, push, or otherwise publish changes: no pull requests, no releases.
- Leave changes in the working tree for the user to review.
