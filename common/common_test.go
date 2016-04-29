package common_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/Orange-OpenSource/travis-resource/common"
	"github.com/Orange-OpenSource/travis-resource/travis"
	"github.com/Orange-OpenSource/travis-resource/model"
)

var _ = Describe("Common", func() {
	Describe("StringInSlice", func() {
		slice := []string{"apple", "orange", "tomato"}
		It("should return true if found", func() {
			Expect(common.StringInSlice("apple", slice)).To(BeTrue())
		})
		It("should return false if not found", func() {
			Expect(common.StringInSlice("kiwi", slice)).To(BeFalse())
		})
	})
	Describe("GetTravisDashboardUrl", func() {
		apiUrl := "https://api.test.com/"
		It("should return only domain name", func() {
			Expect(common.GetTravisDashboardUrl(apiUrl)).To(BeEquivalentTo("https://test.com/"))
		})
	})
	Describe("GetMetadatasFromBuild", func() {

		Context("with build in error", func() {
			build := travis.Build{
				State: "errored",
			}
			expectedMetadata := []model.Metadata{model.Metadata{"travis_succeeded", "false"}, model.Metadata{"travis_build_state", "errored"}}
			It("should give the actual state and travis suceeded to false", func() {
				Expect(common.GetMetadatasFromBuild(build)).To(BeEquivalentTo(expectedMetadata))
			})
		})
		Context("with started build", func() {
			build := travis.Build{
				State: "started",
			}
			expectedMetadata := []model.Metadata{model.Metadata{"travis_succeeded", "false"}, model.Metadata{"travis_build_state", "started"}}
			It("should give the actual state and travis suceeded to false", func() {
				Expect(common.GetMetadatasFromBuild(build)).To(BeEquivalentTo(expectedMetadata))
			})
		})
		Context("with build in success", func() {
			build := travis.Build{
				State: travis.SUCCEEDED_STATE,
			}
			expectedMetadata := []model.Metadata{model.Metadata{"travis_succeeded", "true"}, model.Metadata{"travis_build_state", travis.SUCCEEDED_STATE}}
			It("should give the actual state and travis suceeded to true", func() {
				Expect(common.GetMetadatasFromBuild(build)).To(BeEquivalentTo(expectedMetadata))
			})
		})
	})
	Describe("GetTravisUrl", func() {
		Context("when choose travis pro", func() {
			It("should give the api endpoint for travis pro", func() {
				Expect(common.GetTravisUrl(true)).To(BeEquivalentTo(travis.TRAVIS_API_PRO_URL))
			})
		})
		Context("when choose travis open source projects", func() {
			It("should give the default api endpoint", func() {
				Expect(common.GetTravisUrl(false)).To(BeEquivalentTo(travis.TRAVIS_API_DEFAULT_URL))
			})
		})
	})
	Describe("MakeTravisClient", func() {
		Context("with no token set", func() {
			source := model.Source{}
			It("should return a travis client for travis pro", func() {
				source.Pro = true
				travisClient, err := common.MakeTravisClient(source)
				Expect(err).To(BeNil())
				Expect(travisClient).NotTo(BeNil())
			})
			It("should return a travis client for travis open source projects", func() {
				source.Pro = false
				travisClient, err := common.MakeTravisClient(source)
				Expect(err).To(BeNil())
				Expect(travisClient).NotTo(BeNil())
			})
		})
		Context("with travis token set", func() {
			source := model.Source{
				TravisToken: "mytravistoken",
			}
			It("should return a travis client for travis pro", func() {
				source.Pro = true
				travisClient, err := common.MakeTravisClient(source)
				Expect(err).To(BeNil())
				Expect(travisClient).NotTo(BeNil())
			})
			It("should return a travis client for travis open source projects", func() {
				source.Pro = false
				travisClient, err := common.MakeTravisClient(source)
				Expect(err).To(BeNil())
				Expect(travisClient).NotTo(BeNil())
			})
		})
		Context("with github token set", func() {
			source := model.Source{
				GithubToken: "mytoken",
			}
			It("should return a travis client for travis pro", func() {
				source.Pro = true
				travisClient, err := common.MakeTravisClient(source)
				Expect(err).NotTo(BeNil())
				Expect(travisClient).To(BeNil())
			})
			It("should return a travis client for travis open source projects", func() {
				source.Pro = false
				travisClient, err := common.MakeTravisClient(source)
				Expect(err).NotTo(BeNil())
				Expect(travisClient).To(BeNil())
			})
		})
	})
})
