---
title: "Writing Procedural Documentation"
description: "Guidelines for writing consistent procedural documentation"
weight: 10
---

## Using the Task Archetype

Envoy Gateway provides a standardized archetype for creating procedural documentation. This ensures consistency across all task-oriented pages and makes it easier for new contributors to write technical documentation.

### Creating a New Task Document

To create a new task document using the archetype, use the Hugo `new` command:

```bash
# From the site directory
hugo new --kind task content/en/latest/tasks/your-category/your-new-task.md
```

This will generate a new markdown file with the predefined structure following our documentation standards.

### Task Document Structure

The generated task document follows this structure:

```markdown
---
title: "Your New Task"
description: "This task provides instructions for configuring your new task in Envoy Gateway."
weight: 10
---

This task provides instructions for configuring your new task.

Your New Task allows you to [briefly explain what this feature does and its benefits].


## Installation

Install the necessary resources...

## Verification

Verify the configuration...

## Testing

Test the configuration...

## Troubleshooting

If you encounter issues...

## Clean-Up

Remove the resources...

## Next Steps

Checkout the Developer Guide...
```

### Common Patterns in Envoy Gateway Tasks

Our task documents typically follow these patterns:

1. **Start with a brief introduction** explaining what the feature does and its benefits
2. **Include the Prerequisites section** using the boilerplate shortcode
3. **Installation section** with resource creation using tabpane for different methods (stdin vs file)
4. **Verification section** showing how to check that resources were created correctly
5. **Testing section** with curl examples and expected outputs
6. **Troubleshooting section** for common issues
7. **Clean-Up section** to remove resources created during the task
8. **Next Steps section** linking to additional resources

### Using Shortcodes

The archetype includes several useful shortcodes:



2. **Tabpane** - For showing alternative approaches:
   ```
   {{< tabpane text=true >}}
   {{% tab header="Option 1" %}}
   Content for option 1
   {{% /tab %}}
   {{% tab header="Option 2" %}}
   Content for option 2
   {{% /tab %}}
   {{< /tabpane >}}
   ```

### Best Practices

1. **Be Clear and Concise**:
   - Use simple, direct language
   - Explain what happens in each step
   - Break down complex procedures into logical sections

2. **Include Working Examples**:
   - Ensure all example code works as written
   - Use consistent naming conventions for resources
   - Show both stdin and file-based approaches

3. **Provide Verification Steps**:
   - Always include commands to verify resource creation
   - Explain what to look for in the status and output
   - Include expected response codes and messages

4. **Document Troubleshooting**:
   - Include a troubleshooting section for common issues
   - Show how to check logs and resource status
   - Provide solutions for typical problems

5. **Complete Cleanup**:
   - Ensure all created resources are removed
   - Include commands in the same order as creation
   - Add comments for clarity

### Directory Structure

Place your new task documents in the appropriate category under `site/content/en/latest/tasks/`:

- **traffic/** - For routing, load balancing, and traffic management tasks
- **security/** - For authentication, authorization, and security tasks
- **observability/** - For metrics, logging, and tracing tasks
- **operations/** - For operational tasks like installation, upgrades
- **extensibility/** - For extending Envoy Gateway functionality

If your task doesn't fit these categories, consider if a new category is needed. 