# Lobster Pot

[![Deploy](https://www.herokucdn.com/deploy/button.svg)](https://heroku.com/deploy?template=https://github.com/salesforce/lobster-pot)

## Purpose

The purpose of this software is to scan all code pushed into one or more Github Organisations, to search for secrets, and report to Slack any findings. It has been originally created by [Etienne Stalmans](https://github.com/staaldraad) and is used actively used in various Github organisations under the Salesforce Enterprise plan.

It has been primarily designed to run on Heroku, but can be used on any platform that supports [12factor apps](https://12factor.net/).

## Monitoring of a GitHub Org

The app receives push event notifications from GitHub. Each push is reviewed and the commits within are scanned for possible secrets (such as passwords, AWS secret keys, API tokens etc).

When the scanning reveals findings, the application posts a message to a defined slack channel with the relevant details and triggers a manual review.
Those findings are also stored in the database for stats and reporting purposes.

![Data Flow Diagram](docs/data-flow-diagram.png)

## Setup

See the [docs/configuration](docs/configuration) folder for the specifics.

At least one github organization and one slack app must be configured for the app to start properly.
