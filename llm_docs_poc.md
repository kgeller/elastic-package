# How to run

1. First you must authenticate with AWS. This can be achieved by running `aws-mfa --profile=elastic-siem`
1. Build the current branch `make build`
1. Install the built version `make install`
1. Select a sample package to run this for (we tested the most on bitwarden) and navigate to the corresponding root directory
1. Update the `_dev/build/docs/README.md` to include `{{ section "overview" }}` and `{{ section "setup" }}`. 
1. Run `elastic-package llm-write-docs`
1. You'll now be able to see the generated content under ` /_dev/build/docs/sections/`.
1. Run `elastic-package build` and the sections will be included in the `docs/README.md`.

Note: you're also able to use the `elastic-package create package` flow to generate the through one of the command prompted questions.

## Sample output from my test run on 2025-05-27

<details>
  <summary>overview</summary>
  
  ```md
The Bitwarden integration enables you to collect and analyze event logs from your Bitwarden organization, providing visibility into password manager usage, security events, and administrative activities across your enterprise.

Bitwarden is a comprehensive password management solution that helps organizations secure credentials, store sensitive information, and enforce password policies for teams and individuals.

This integration captures audit logs and security events from Bitwarden's API, allowing you to monitor user authentication, vault access, policy violations, and administrative changes within your password management infrastructure.

By ingesting Bitwarden logs into Elastic, you can:
- Correlate password manager events with other security data
- Detect suspicious authentication patterns
- Track compliance with password policies
- Gain insights into credential management practices across your organization

The integration supports monitoring activities such as:
- User logins
- Vault item access
- Sharing events
- Policy enforcement
- Administrative actions

**Example use cases:**

You can identify users who frequently access shared credentials, detect unusual login patterns that might indicate compromised accounts, or track compliance with password rotation policies by analyzing vault modification events alongside your broader security monitoring strategy.


<!-- Generated on: 2025-05-27T13:34:50-04:00 -->
  ```

</details>

<details>
  <summary>setup</summary>
  
  ```md
To configure Bitwarden for log collection, you need to set up API access and enable event logging in your Bitwarden organization.

**Prerequisites:**
- You must have an active Bitwarden organization with administrative privileges.

**Setup instructions:**
- 1. Log in to your Bitwarden organization as an administrator.
- 2. Navigate to the "Settings" section and select "API Keys".
- 3. Create a new API key with the appropriate permissions for accessing event logs (e.g., "Access.EventLogs").
- 4. Copy the generated API key for use in the Elastic integration.
- 5. In the Bitwarden settings, enable event logging by navigating to "Audit Log" and toggling the "Enable Audit Log" option.

For detailed instructions on configuring Bitwarden, refer to the official Bitwarden documentation: https://bitwarden.com/help/article/audit-log/


<!-- Generated on: 2025-05-27T13:34:50-04:00 -->
  ```
  
</details>