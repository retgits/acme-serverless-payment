# Continuous Verification using GitHub Actions

> A process of querying external systems and using information from the response to make decisions to improve the development and deployment process.

Continuous Verification, by default, means it’s an extension to the existing development and deployment processes that companies already have. That extension focuses on optimizing both the development and deployment experience for those companies by looking at security, performance, and cost. Rather than doing that manually, you can do it in your existing CI/CD pipeline, with GitHub actions.

To make this sample work, you'll need to host your own version of the GitHub Actions runner or change `runs-on` in [`cv.yaml`](./cv.yaml). There are a few external services you'll need access to as well that are configured using "secrets":

* `${{ secrets.SNYK_TOKEN }}`: To check for vulnerabilities this workflow uses Snyk. You'll need an API Access Token to be able to connect to Snyk from the workflow and see whether there are any vulnerabilities
* `${{ secrets.AWS_ACCESS_KEY_ID }}`: The AWS Access key used to deploy the service to AWS Lambda
* `${{ secrets.AWS_SECRET_ACCESS_KEY }}`: The AWS Secret key used to deploy the service to AWS Lambda

## More information

[Continuous Verification: The Missing Link to Fully Automate Your Pipeline](https://thenewstack.io/continuous-verification-the-missing-link-to-fully-automate-your-pipeline/)