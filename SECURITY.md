# Security Policy

## Supported Versions

We actively support the following versions of Fabricator with security updates:

| Version | Supported          |
| ------- | ------------------ |
| 1.x.x   | :white_check_mark: |
| < 1.0   | :x:                |

## Reporting a Vulnerability

We take security vulnerabilities seriously. If you discover a security vulnerability, please follow these steps:

### 🔒 Private Disclosure

**DO NOT** open a public issue for security vulnerabilities.

Instead, please report security vulnerabilities through one of these channels:

1. **GitHub Security Advisories** (Preferred)
   - Go to the [Security tab](https://github.com/SGNL-ai/fabricator/security) of this repository
   - Click "Report a vulnerability"
   - Provide detailed information about the vulnerability

2. **Email** (Alternative)
   - Send details to: [security@sgnl.ai](mailto:security@sgnl.ai)
   - Include "Fabricator Security Vulnerability" in the subject line

### 📋 What to Include

When reporting a vulnerability, please include:

- **Description** of the vulnerability
- **Steps to reproduce** the issue
- **Potential impact** of the vulnerability
- **Suggested fix** (if you have one)
- **Your contact information** for follow-up questions

### 🕐 Response Timeline

- **Initial Response**: Within 48 hours
- **Detailed Assessment**: Within 7 days
- **Fix Timeline**: Depends on severity (24 hours for critical, 30 days for low)

## Security Measures

### 🛡️ Built-in Security Features

- **Input Validation**: Comprehensive YAML schema validation
- **Path Sanitization**: Safe file path handling for CSV output
- **Resource Limits**: Bounded memory usage and file operations
- **No Network Calls**: Purely local operation, no external dependencies at runtime

### 🔍 Automated Security Scanning

Our CI/CD pipeline includes:

- **CodeQL Analysis**: GitHub's semantic code analysis
- **gosec**: Go security analyzer for common vulnerabilities
- **govulncheck**: Known vulnerability detection in dependencies
- **Dependabot**: Automated dependency vulnerability updates

### 🏗️ Secure Development Practices

- **Least Privilege**: GitHub Actions use minimal required permissions
- **Supply Chain Security**: All dependencies tracked and auto-updated
- **Code Review**: All changes require review before merge
- **Automated Testing**: Comprehensive test suite prevents regressions

## Security Considerations for Users

### 🔐 Safe Usage

- **File Permissions**: Ensure output directories have appropriate permissions
- **Input Validation**: Only use trusted YAML definition files
- **Resource Monitoring**: Monitor disk space when generating large datasets

### ⚠️ Potential Risks

- **Large File Generation**: Be cautious with high row counts to avoid disk space issues
- **YAML Complexity**: Very complex YAML files may consume significant memory
- **Output Directory**: Tool writes files to specified directory - ensure it's safe

### 🛠️ Secure Configuration

```bash
# Good: Use explicit output directory
./fabricator -f template.yaml -o ./safe-output-dir/

# Good: Limit data volume for testing
./fabricator -f template.yaml -n 100 -o ./output/

# Caution: Very large datasets
./fabricator -f template.yaml -n 1000000 -o ./output/  # Monitor disk space
```

## Vulnerability Disclosure History

No security vulnerabilities have been reported to date.

---

**Questions about security?** Contact us at [security@sgnl.ai](mailto:security@sgnl.ai)