package isolated

import (
	"sort"

	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("org command", func() {
	var (
		orgName   string
		spaceName string
	)

	BeforeEach(func() {
		orgName = helpers.NewOrgName()
		spaceName = helpers.PrefixedRandomName("SPACE")
	})

	Describe("help", func() {
		Context("when --help flag is set", func() {
			It("Displays command usage to output", func() {
				session := helpers.CF("org", "--help")
				Eventually(session).Should(Say("NAME:"))
				Eventually(session).Should(Say("org - Show org info"))
				Eventually(session).Should(Say("USAGE:"))
				Eventually(session).Should(Say("cf org ORG [--guid]"))
				Eventually(session).Should(Say("OPTIONS:"))
				Eventually(session).Should(Say("--guid\\s+Retrieve and display the given org's guid.  All other output for the org is suppressed."))
				Eventually(session).Should(Say("SEE ALSO:"))
				Eventually(session).Should(Say("org-users, orgs"))
				Eventually(session).Should(Exit(0))
			})
		})
	})

	Context("when the environment is not setup correctly", func() {
		Context("when no API endpoint is set", func() {
			BeforeEach(func() {
				helpers.UnsetAPI()
			})

			It("fails with no API endpoint set message", func() {
				session := helpers.CF("org", orgName)
				Eventually(session.Out).Should(Say("FAILED"))
				Eventually(session.Out).Should(Say("No API endpoint set. Use 'cf login' or 'cf api' to target an endpoint."))
				Eventually(session).Should(Exit(1))
			})
		})

		Context("when not logged in", func() {
			BeforeEach(func() {
				helpers.LogoutCF()
			})

			It("fails with not logged in message", func() {
				session := helpers.CF("org", orgName)
				Eventually(session.Out).Should(Say("FAILED"))
				Eventually(session.Out).Should(Say("Not logged in. Use 'cf login' to log in."))
				Eventually(session).Should(Exit(1))
			})
		})
	})

	Context("when the environment is set up correctly", func() {
		BeforeEach(func() {
			helpers.LoginCF()
		})

		Context("when the org does not exist", func() {
			It("displays org not found and exits 1", func() {
				session := helpers.CF("org", orgName)
				Eventually(session.Out).Should(Say("FAILED"))
				Eventually(session.Out).Should(Say("Organization %s not found", orgName))
				Eventually(session).Should(Exit(1))
			})
		})

		Context("when the org exists", func() {
			BeforeEach(func() {
				setupCF(orgName, spaceName)
			})

			Context("when the --guid flag is used", func() {
				It("displays the org guid", func() {
					session := helpers.CF("org", "--guid", orgName)
					Eventually(session.Out).Should(Say("[\\da-f]{8}-[\\da-f]{4}-[\\da-f]{4}-[\\da-f]{4}-[\\da-f]{12}"))
					Eventually(session).Should(Exit(0))
				})
			})

			Context("when no flags are used", func() {
				var (
					domainName      string
					quotaName       string
					spaceName2      string
					spaceQuotaName  string
					spaceQuotaName2 string
				)

				BeforeEach(func() {
					domainName = helpers.DomainName("")
					domain := helpers.NewDomain(orgName, domainName)
					domain.Create()

					quotaName = helpers.QuotaName()
					session := helpers.CF("create-quota", quotaName, "-a", "654", "--allow-paid-service-plans", "-i", "456M", "-m", "123M", "-r", "789", "--reserved-route-ports", "321", "-s", "987")
					Eventually(session).Should(Exit(0))
					session = helpers.CF("set-quota", orgName, quotaName)
					Eventually(session).Should(Exit(0))

					spaceName2 = helpers.PrefixedRandomName("SPACE")
					helpers.CreateSpace(spaceName2)

					spaceQuotaName = helpers.QuotaName()
					session = helpers.CF("create-space-quota", spaceQuotaName)
					Eventually(session).Should(Exit(0))
					session = helpers.CF("set-space-quota", spaceName, spaceQuotaName)
					Eventually(session).Should(Exit(0))

					spaceQuotaName2 = helpers.QuotaName()
					session = helpers.CF("create-space-quota", spaceQuotaName2)
					Eventually(session).Should(Exit(0))
					session = helpers.CF("set-space-quota", spaceName2, spaceQuotaName2)
					Eventually(session).Should(Exit(0))
				})

				It("displays a table with org domains, quotas, spaces and space quotas and exits 0", func() {
					session := helpers.CF("org", orgName)
					userName, _ := helpers.GetCredentials()
					Eventually(session.Out).Should(Say("Getting info for org %s as %s...", orgName, userName))
					Eventually(session.Out).Should(Say("OK"))

					Eventually(session.Out).Should(Say("%s:", orgName))

					domainsSorted := []string{defaultSharedDomain(), domainName}
					sort.Strings(domainsSorted)
					Eventually(session.Out).Should(Say("domains:\\s+%s, %s", domainsSorted[0], domainsSorted[1]))

					Eventually(session.Out).Should(Say("quota:\\s+%s \\(123M memory limit, 456M instance memory limit, 789 routes, 987 services, paid services allowed, 654 app instance limit, 321 route ports\\)", quotaName))

					spacesSorted := []string{spaceName, spaceName2}
					sort.Strings(spacesSorted)
					Eventually(session.Out).Should(Say("spaces:\\s+%s, %s", spacesSorted[0], spacesSorted[1]))

					spaceQuotasSorted := []string{spaceQuotaName, spaceQuotaName2}
					sort.Strings(spaceQuotasSorted)
					Eventually(session.Out).Should(Say("space quotas:\\s+%s, %s", spaceQuotasSorted[0], spaceQuotasSorted[1]))

					Eventually(session).Should(Exit(0))
				})

				Context("when quota setting are defaults or unlimited", func() {
					BeforeEach(func() {
						quotaName = helpers.QuotaName()
						session := helpers.CF("create-quota", quotaName, "--reserved-route-ports", "-1")
						Eventually(session).Should(Exit(0))
						session = helpers.CF("set-quota", orgName, quotaName)
						Eventually(session).Should(Exit(0))
					})

					It("displays a table with org domains, quotas, spaces and space quotas and exits 0", func() {
						session := helpers.CF("org", orgName)
						Eventually(session.Out).Should(Say("quota:\\s+%s \\(0M memory limit, unlimited instance memory limit, 0 routes, 0 services, paid services disallowed, unlimited app instance limit, unlimited route ports\\)", quotaName))
						Eventually(session).Should(Exit(0))
					})
				})
			})
		})
	})
})
