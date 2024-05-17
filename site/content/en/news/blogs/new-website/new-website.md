---
date: 2023-10-08
title: Welcome to new website!
linkTitle: Welcome to new website
description: >
  We migrate our docs from Sphinx to Hugo now!
author: Xunzhuo Liu
---

{{% alert title="Summary" color="primary" %}}
Migrate from ***Sphinx*** to ***Hugo*** for Envoy Gateway Documents.
{{% /alert %}}

## Introduction

In the realm of static site generators, two names often come up: Sphinx and Hugo. While both are powerful tools, we recently made the decision to migrate our documentation from Sphinx to Hugo. This article aims to shed light on the reasons behind this move and the advantages we've discovered in using Hugo for static blogging.

## Why Migrate?

Sphinx, originally created for Python documentation, has served us well over the years. It offers a robust and flexible solution for technical documentation. However, as our needs evolved, we found ourselves seeking a tool that could offer faster build times, ease of use, and a more dynamic community. This led us to Hugo.

## Advantages of Hugo

+ **Speed**: Hugo is renowned for its speed. It can build a site in a fraction of the time it takes other static site generators. This is a significant advantage when working with large sites or when quick updates are necessary.

+ **Ease of Use**: Hugo's simplicity is another strong point. It doesn't require a runtime environment, and its installation process is straightforward. Moreover, Hugo's content management is intuitive, making it easy for non-technical users to create and update content.

+ **Flexibility**: Hugo supports a wide range of content types, from blog posts to documentation. It also allows for custom outputs, enabling us to tailor our site to our specific needs.

+ **Active Community**: Hugo boasts a vibrant and active community. This means regular updates, a wealth of shared themes and plugins, and a responsive support network.

+ **Multilingual Support**: Hugo's built-in support for multiple languages is a boon for global teams. It allows us to create content in various languages without the need for additional plugins or tools.

+ **Markdown Support**: Hugo's native support for Markdown makes it easy to write and format content. This is particularly beneficial for technical writing, where code snippets and technical formatting are common.

## Challenges Encountered During the Migration

While the migration from Sphinx to Hugo has brought numerous benefits, it was not without its challenges and it really took me a lot of time on it. Here are some of the difficulties we encountered during the process:

+ **Converting RST to Markdown**: Our documentation contained a large number of reStructuredText (RST) files, which needed to be converted to Markdown, the format Hugo uses. This required a careful and meticulous conversion process to ensure no information was lost or incorrectly formatted.

+ **Adding Headings to tons of Markdown Files**: Hugo requires headings in its Markdown files, which our old documents did not have. We had to write scripts to add these headings in bulk, a task that required a deep understanding of both our content and Hugo's requirements.

+ **Handling Multiple Versions**: Our documentation had already gone through five iterations, resulting in a large number of files to manage. We had to ensure that all versions were correctly migrated and that the versioning system in Hugo was correctly set up.

+ **Designing a New Page Structure and Presentation**: To provide a better reading experience, we needed to design a new way to organize and present our pages. This involved understanding our readers' needs and how best to structure our content to meet those needs.

+ **Updating Existing Toolchains**: The migration also required us to update our existing toolchains, including Makefile, CI, release processes, and auto-generation tools. This was a complex task that required a deep understanding of both our old and new systems.

Despite these challenges, the benefits of migrating to Hugo have far outweighed the difficulties. The process, while complex, has provided us with a more efficient, user-friendly, and flexible system for managing our static blog. It's a testament to the power of Hugo and the value of continuous improvement in the tech world.

## Conclusion

While Sphinx has its strengths, our migration to Hugo has opened up new possibilities for our static blogging. The speed, ease of use, flexibility, and active community offered by Hugo have made it a powerful tool in our arsenal.

## Some Words

{{% alert title="Author" color="primary" %}}
[Xunzhuo Liu](https://github.com/Xunzhuo), the maintainer and steering committee member of Envoy Gateway.
{{% /alert %}}

From the inception of my involvement in the project, I have placed great emphasis on user experience, including both developer and end-user experiences. I have built a rich toolchain, automated pipelines, Helm Charts, command-line tools, and documentation to enhance the overall experience of interacting with the project.

My dedication to improving user experience was a driving force behind the decision to migrate from Sphinx to Hugo. I recognized the potential for Hugo to provide a more intuitive and efficient platform for managing the project's static blog. Despite the challenges encountered during the migration, my commitment to enhancing user experience ensured a successful transition.

Through this migration, I have further demonstrated my commitment to continuous improvement and my ability to adapt to the evolving needs of the project. My work serves as a testament to the importance of user experience in software development and the value of embracing new tools and technologies to meet these needs.

Enjoy new Envoy Gateway Website! ❤️
