<p align="center">
	<a href="https://writefreely.org"><img src="https://writefreely.org/img/logo.png" width="200px" alt="WriteFreely" /></a>
</p>

# WriteFreely

WriteFreely is a clean, minimalist publishing platform made for writers. Start a blog, share knowledge within your organization, or build a community around the shared act of writing.

## This fork

Welcome to my fork of writefreely!

We've got:

+ NO monetization
+ NO email
+ NO user invites
+ NO user signup
+ NO third-party OAuth
+ NO spurious tags
+ Fixes for single-user sites
+ Search functionality
+ Improved localization
+ Working archive with tags
+ More configuration options
+ Improved draft editing
+ Markdown parsing with [gomarkdown](https://github.com/gomarkdown/markdown)
+ Unobtrusive table of contents (based on [kjk's](https://gist.github.com/kjk/d9343c3f45d9f529b2b8156048254840))

## Features

### Made for writing

Built on a plain, auto-saving editor, WriteFreely gives you a distraction-free writing environment. Once published, your words are front and center, and easy to read.

### A connected community

Start writing together, publicly or privately. Connect with other communities, whether running WriteFreely, [Plume](https://joinplu.me/), or other ActivityPub-powered software. And bring members on board from your existing platforms, thanks to our OAuth 2.0 support.

### Intuitive organization

Categorize articles [with hashtags](https://writefreely.org/docs/latest/writer/hashtags), and create static pages from normal posts by [_pinning_ them](https://writefreely.org/docs/latest/writer/static) to your blog. Create draft posts and publish to multiple blogs from one account.

### International

Blog elements are localized in 20+ languages, and WriteFreely includes first-class support for non-Latin and right-to-left (RTL) script languages.

### Private by default

WriteFreely collects minimal data, and never publicizes more than a writer consents to. Writers can seamlessly create multiple blogs from a single account for different pen names or purposes without publicly revealing their association.

<h2><a href="https://write.as/writefreely"><img src="https://writefreely.org/img/writeas-readme.png" height="32px" alt="Write.as" /></a></h2>

The quickest way to deploy WriteFreely is with [Write.as](https://write.as/writefreely), a hosted service from the team behind WriteFreely. You'll get fully-managed installation, backup, upgrades, and maintenance — and directly fund our free software work ❤️

[**Learn more on Write.as**](https://write.as/writefreely).

## Quick start

WriteFreely deploys as a static binary on any platform and architecture that Go supports. Just use our built-in SQLite support, or add a MySQL or MariaDB database, and you'll be up and running!

For common platforms, start with our [pre-built binaries](https://github.com/writefreely/writefreely/releases/) and head over to our [installation guide](https://writefreely.org/start) to get started.

### Packages

You can also find WriteFreely in these package repositories, thanks to our wonderful community!

* [Arch User Repository](https://aur.archlinux.org/packages/writefreely/)
* [Nanos Repository](https://repo.ops.city/v2/packages/eyberg/writefreely/show)

## Documentation

Read our full [documentation on WriteFreely.org](https://writefreely.org/docs) &mdash;️ and help us improve by contributing to the [writefreely/documentation](https://github.com/writefreely/documentation) repo.

## Development

Start hacking on WriteFreely with our [developer setup guide](https://writefreely.org/docs/latest/developer/setup). For Docker support, see our [Docker guide](https://writefreely.org/docs/latest/admin/docker).

## Contributing

We gladly welcome contributions to WriteFreely, whether in the form of [code](https://github.com/writefreely/writefreely/blob/master/CONTRIBUTING.md#contributing-to-writefreely), [bug reports](https://github.com/writefreely/writefreely/issues/new?template=bug_report.md), [feature requests](https://discuss.write.as/c/feedback/feature-requests), [translations](https://poeditor.com/join/project/TIZ6HFRFdE), or [documentation](https://github.com/writefreely/documentation) improvements.

Before contributing anything, please read our [Contributing Guide](https://github.com/writefreely/writefreely/blob/master/CONTRIBUTING.md#contributing-to-writefreely). It describes the correct channels for submitting contributions and any potential requirements.

## License

Copyright © 2018-2026 [Musing Studio LLC](https://musing.studio) and contributing authors. Licensed under the [AGPL](https://github.com/writefreely/writefreely/blob/develop/LICENSE).
