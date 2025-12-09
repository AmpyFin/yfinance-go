# Security Policy

## Supported Versions

We provide security updates for the following versions:

| Version | Supported          |
| ------- | ------------------ |
| 1.1.x   | :white_check_mark: |
| 1.0.x   | :white_check_mark: |
| < 1.0   | :x:                |

## Reporting a Vulnerability

**Please do not report security vulnerabilities through public GitHub issues.**

Instead, please report them via one of the following methods:

### Preferred Method: Email

Send an email to: **security@ampyfin.com**

Include the following information:
- Type of vulnerability
- Affected component(s)
- Steps to reproduce
- Potential impact
- Suggested fix (if any)

### Alternative: Private Security Advisory

If you have a GitHub account, you can create a private security advisory:
1. Go to the repository's Security tab
2. Click "Report a vulnerability"
3. Fill out the security advisory form

## Response Timeline

- **Initial Response**: Within 48 hours
- **Status Update**: Within 7 days
- **Resolution**: Depends on severity and complexity

## Security Best Practices

### For Users

1. **Keep dependencies updated**:
   ```bash
   go get -u ./...
   go mod tidy
   ```

2. **Use latest version** of yfinance-go

3. **Review configuration** for sensitive settings

4. **Monitor for security advisories**

### For Developers

1. **Never commit secrets** (API keys, tokens, passwords)

2. **Use environment variables** for sensitive configuration

3. **Validate all inputs** from external sources

4. **Follow secure coding practices**:
   - Avoid SQL injection (N/A for this project)
   - Sanitize user inputs
   - Use parameterized queries
   - Implement proper error handling

5. **Keep dependencies updated**:
   ```bash
   go get -u all
   go mod tidy
   ```

## Known Security Considerations

### Rate Limiting

- The library implements rate limiting to prevent abuse
- Configure appropriate QPS limits for your use case
- Monitor for rate limit violations

### Web Scraping

- The library respects `robots.txt` by default
- Use scraping responsibly and in accordance with Yahoo Finance's terms of service
- Implement appropriate delays and backoff strategies

### Network Security

- All network requests use HTTPS
- Certificate validation is enabled by default
- Do not disable TLS verification in production

### Data Validation

- All data from external sources is validated
- Protobuf schemas provide type safety
- Input validation prevents injection attacks

## Security Updates

Security updates will be:
- Released as patch versions (e.g., 1.1.1 → 1.1.2)
- Documented in CHANGELOG.md
- Tagged with security labels on GitHub

## Disclosure Policy

1. **Private Disclosure**: We will work with you to fix the vulnerability privately
2. **Coordinated Disclosure**: We will coordinate public disclosure after a fix is available
3. **Credit**: We will credit you (if desired) in the security advisory

## Security Checklist for Contributors

Before submitting code, ensure:

- [ ] No hardcoded secrets or credentials
- [ ] Input validation for all user-provided data
- [ ] Proper error handling (no information leakage)
- [ ] Dependencies are up to date
- [ ] No use of deprecated or insecure functions
- [ ] Proper use of context for cancellation/timeouts
- [ ] Rate limiting is respected
- [ ] No sensitive data in logs

## Security Audit

We periodically:
- Review dependencies for known vulnerabilities
- Audit code for security issues
- Update security best practices
- Review and update this policy

## Questions?

For security-related questions that are not vulnerabilities:
- Open a GitHub Discussion
- Check existing documentation
- Review GitHub Issues (non-security)

---

**Thank you for helping keep yfinance-go secure!**

