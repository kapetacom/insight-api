# insight-api
API Service for accessing runtime data

# Install Air for auto reload when doing changes

`go install github.com/cosmtrek/air@latest`

# Releasing the insight-api
To release the insight-api, follow these steps:

1. Create a new tag in GitHub for the release. The tag should be in the format vX.Y.Z, where X, Y, and Z are the major, minor, and patch versions of the release, respectively.
2. Push the tag to GitHub.
3. Update the deployment-target-go with the new tag. To do this, edit the [insightsDeployment.yml](https://github.com/kapetacom/deployment-target-gcp-go/blob/427d013af3bd90b0fefc4ef04b92c860e96c3361/pkg/templates/insightsDeployment.yml#L32) file and replace the old tag with the new tag.
4. Commit and push the changes to the deployment-target-go repository.

The new release will be automatically deployed next time people are deploying.