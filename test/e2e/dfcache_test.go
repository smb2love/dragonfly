/*
 *     Copyright 2025 The Dragonfly Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package e2e

import (
	"fmt"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2" //nolint
	. "github.com/onsi/gomega"    //nolint

	"d7y.io/dragonfly/v2/pkg/idgen"
	"d7y.io/dragonfly/v2/test/e2e/util"
)

func dfcacheImportWithRetry(clientPod *util.PodExec, cmd string) ([]byte, error) {
	var (
		out []byte
		err error
	)

	// The client host can be temporarily missing in persistent-cache host manager (redis),
	// e.g. during (re-)announces. Retry only this known transient condition.
	deadline := time.Now().Add(30 * time.Second)
	for {
		out, err = clientPod.Command("sh", "-c", cmd).CombinedOutput()
		if err == nil {
			return out, nil
		}

		outStr := string(out)
		if strings.Contains(outStr, "host ") && strings.Contains(outStr, " not found") && time.Now().Before(deadline) {
			time.Sleep(2 * time.Second)
			continue
		}

		return out, err
	}
}

var _ = Describe("Import and Export Using Dfcache", func() {
	Context("1MiB file", func() {
		var (
			testFile *util.File
			err      error
		)

		BeforeEach(func() {
			testFile, err = util.GetFileServer().GenerateFile(util.FileSize1MiB)
			Expect(err).NotTo(HaveOccurred())
			Expect(testFile).NotTo(BeNil())
		})

		AfterEach(func() {
			err = util.GetFileServer().DeleteFile(testFile.GetInfo())
			Expect(err).NotTo(HaveOccurred())
		})

		It("import and export should be ok", Label("dfcache", "import and export"), func() {
			var clientPod *util.PodExec
			clientPod, err = util.ClientExec()
			fmt.Println(err)
			Expect(err).NotTo(HaveOccurred())

			_, err := clientPod.Command("sh", "-c", fmt.Sprintf("dfget %s --disable-back-to-source --output %s", testFile.GetDownloadURL(), testFile.GetOutputPath())).CombinedOutput()
			Expect(err).NotTo(HaveOccurred())

			sha256sum, err := util.CalculateSha256ByTaskID([]*util.PodExec{clientPod}, testFile.GetTaskID())
			Expect(err).NotTo(HaveOccurred())
			Expect(testFile.GetSha256()).To(Equal(sha256sum))

			sha256sum, err = util.CalculateSha256ByOutput([]*util.PodExec{clientPod}, testFile.GetOutputPath())
			Expect(err).NotTo(HaveOccurred())
			Expect(testFile.GetSha256()).To(Equal(sha256sum))

			importOut, err := dfcacheImportWithRetry(clientPod, fmt.Sprintf("dfcache import %s --persistent-replica-count 1 --content-for-calculating-task-id %s", testFile.GetOutputPath(), testFile.GetSha256()))
			fmt.Println(string(importOut), err)
			Expect(err).NotTo(HaveOccurred())

			seedClientPod, err := util.SeedClientExec(0)
			fmt.Println(err)
			Expect(err).NotTo(HaveOccurred())

			taskID := idgen.PersistentCacheTaskIDByContent(testFile.GetSha256())
			exportOut, err := seedClientPod.Command("sh", "-c", fmt.Sprintf("dfcache export %s --output %s", taskID, testFile.GetOutputPath())).CombinedOutput()
			fmt.Println(string(exportOut), err)
			Expect(err).NotTo(HaveOccurred())

			sha256sum, err = util.CalculateSha256ByPersistentCacheTaskID([]*util.PodExec{seedClientPod}, taskID)
			Expect(err).NotTo(HaveOccurred())
			Expect(testFile.GetSha256()).To(Equal(sha256sum))

			sha256sum, err = util.CalculateSha256ByOutput([]*util.PodExec{seedClientPod}, testFile.GetOutputPath())
			Expect(err).NotTo(HaveOccurred())
			Expect(testFile.GetSha256()).To(Equal(sha256sum))
		})
	})

	Context("10MiB file", func() {
		var (
			testFile *util.File
			err      error
		)

		BeforeEach(func() {
			testFile, err = util.GetFileServer().GenerateFile(util.FileSize10MiB)
			Expect(err).NotTo(HaveOccurred())
			Expect(testFile).NotTo(BeNil())
		})

		AfterEach(func() {
			err = util.GetFileServer().DeleteFile(testFile.GetInfo())
			Expect(err).NotTo(HaveOccurred())
		})

		It("import and export should be ok", Label("dfcache", "import and export"), func() {
			var clientPod *util.PodExec
			clientPod, err = util.ClientExec()
			fmt.Println(err)
			Expect(err).NotTo(HaveOccurred())

			_, err := clientPod.Command("sh", "-c", fmt.Sprintf("dfget %s --disable-back-to-source --output %s", testFile.GetDownloadURL(), testFile.GetOutputPath())).CombinedOutput()
			Expect(err).NotTo(HaveOccurred())

			sha256sum, err := util.CalculateSha256ByTaskID([]*util.PodExec{clientPod}, testFile.GetTaskID())
			Expect(err).NotTo(HaveOccurred())
			Expect(testFile.GetSha256()).To(Equal(sha256sum))

			sha256sum, err = util.CalculateSha256ByOutput([]*util.PodExec{clientPod}, testFile.GetOutputPath())
			Expect(err).NotTo(HaveOccurred())
			Expect(testFile.GetSha256()).To(Equal(sha256sum))

			importOut, err := dfcacheImportWithRetry(clientPod, fmt.Sprintf("dfcache import %s --persistent-replica-count 1 --content-for-calculating-task-id %s", testFile.GetOutputPath(), testFile.GetSha256()))
			fmt.Println(string(importOut), err)
			Expect(err).NotTo(HaveOccurred())

			seedClientPod, err := util.SeedClientExec(0)
			fmt.Println(err)
			Expect(err).NotTo(HaveOccurred())

			taskID := idgen.PersistentCacheTaskIDByContent(testFile.GetSha256())
			exportOut, err := seedClientPod.Command("sh", "-c", fmt.Sprintf("dfcache export %s --output %s", taskID, testFile.GetOutputPath())).CombinedOutput()
			fmt.Println(string(exportOut), err)
			Expect(err).NotTo(HaveOccurred())

			sha256sum, err = util.CalculateSha256ByPersistentCacheTaskID([]*util.PodExec{seedClientPod}, taskID)
			Expect(err).NotTo(HaveOccurred())
			Expect(testFile.GetSha256()).To(Equal(sha256sum))

			sha256sum, err = util.CalculateSha256ByOutput([]*util.PodExec{seedClientPod}, testFile.GetOutputPath())
			Expect(err).NotTo(HaveOccurred())
			Expect(testFile.GetSha256()).To(Equal(sha256sum))
		})
	})

	Context("100MiB file", func() {
		var (
			testFile *util.File
			err      error
		)

		BeforeEach(func() {
			testFile, err = util.GetFileServer().GenerateFile(util.FileSize100MiB)
			Expect(err).NotTo(HaveOccurred())
			Expect(testFile).NotTo(BeNil())
		})

		AfterEach(func() {
			err = util.GetFileServer().DeleteFile(testFile.GetInfo())
			Expect(err).NotTo(HaveOccurred())
		})

		It("import and export should be ok", Label("dfcache", "import and export"), func() {
			var clientPod *util.PodExec
			clientPod, err = util.ClientExec()
			fmt.Println(err)
			Expect(err).NotTo(HaveOccurred())

			_, err := clientPod.Command("sh", "-c", fmt.Sprintf("dfget %s --disable-back-to-source --output %s", testFile.GetDownloadURL(), testFile.GetOutputPath())).CombinedOutput()
			Expect(err).NotTo(HaveOccurred())

			sha256sum, err := util.CalculateSha256ByTaskID([]*util.PodExec{clientPod}, testFile.GetTaskID())
			Expect(err).NotTo(HaveOccurred())
			Expect(testFile.GetSha256()).To(Equal(sha256sum))

			sha256sum, err = util.CalculateSha256ByOutput([]*util.PodExec{clientPod}, testFile.GetOutputPath())
			Expect(err).NotTo(HaveOccurred())
			Expect(testFile.GetSha256()).To(Equal(sha256sum))

			importOut, err := dfcacheImportWithRetry(clientPod, fmt.Sprintf("dfcache import %s --persistent-replica-count 1 --content-for-calculating-task-id %s", testFile.GetOutputPath(), testFile.GetSha256()))
			fmt.Println(string(importOut), err)
			Expect(err).NotTo(HaveOccurred())

			seedClientPod, err := util.SeedClientExec(0)
			fmt.Println(err)
			Expect(err).NotTo(HaveOccurred())

			taskID := idgen.PersistentCacheTaskIDByContent(testFile.GetSha256())
			exportOut, err := seedClientPod.Command("sh", "-c", fmt.Sprintf("dfcache export %s --output %s", taskID, testFile.GetOutputPath())).CombinedOutput()
			fmt.Println(string(exportOut), err)
			Expect(err).NotTo(HaveOccurred())

			sha256sum, err = util.CalculateSha256ByPersistentCacheTaskID([]*util.PodExec{seedClientPod}, taskID)
			Expect(err).NotTo(HaveOccurred())
			Expect(testFile.GetSha256()).To(Equal(sha256sum))

			sha256sum, err = util.CalculateSha256ByOutput([]*util.PodExec{seedClientPod}, testFile.GetOutputPath())
			Expect(err).NotTo(HaveOccurred())
			Expect(testFile.GetSha256()).To(Equal(sha256sum))
		})
	})

	Context("1MiB file and set export transfer-from-dfdaemon", func() {
		var (
			testFile *util.File
			err      error
		)

		BeforeEach(func() {
			testFile, err = util.GetFileServer().GenerateFile(util.FileSize1MiB)
			Expect(err).NotTo(HaveOccurred())
			Expect(testFile).NotTo(BeNil())
		})

		AfterEach(func() {
			err = util.GetFileServer().DeleteFile(testFile.GetInfo())
			Expect(err).NotTo(HaveOccurred())
		})

		It("import and export should be ok", Label("dfcache", "import and export", "export transfer-from-dfdaemon"), func() {
			var clientPod *util.PodExec
			clientPod, err = util.ClientExec()
			fmt.Println(err)
			Expect(err).NotTo(HaveOccurred())

			_, err := clientPod.Command("sh", "-c", fmt.Sprintf("dfget %s --disable-back-to-source --output %s", testFile.GetDownloadURL(), testFile.GetOutputPath())).CombinedOutput()
			Expect(err).NotTo(HaveOccurred())

			sha256sum, err := util.CalculateSha256ByTaskID([]*util.PodExec{clientPod}, testFile.GetTaskID())
			Expect(err).NotTo(HaveOccurred())
			Expect(testFile.GetSha256()).To(Equal(sha256sum))

			sha256sum, err = util.CalculateSha256ByOutput([]*util.PodExec{clientPod}, testFile.GetOutputPath())
			Expect(err).NotTo(HaveOccurred())
			Expect(testFile.GetSha256()).To(Equal(sha256sum))

			importOut, err := dfcacheImportWithRetry(clientPod, fmt.Sprintf("dfcache import %s --persistent-replica-count 1 --content-for-calculating-task-id %s", testFile.GetOutputPath(), testFile.GetSha256()))
			fmt.Println(string(importOut), err)
			Expect(err).NotTo(HaveOccurred())

			seedClientPod, err := util.SeedClientExec(0)
			fmt.Println(err)
			Expect(err).NotTo(HaveOccurred())

			taskID := idgen.PersistentCacheTaskIDByContent(testFile.GetSha256())
			exportOut, err := seedClientPod.Command("sh", "-c", fmt.Sprintf("dfcache export %s --transfer-from-dfdaemon --output %s", taskID, testFile.GetOutputPath())).CombinedOutput()
			fmt.Println(string(exportOut), err)
			Expect(err).NotTo(HaveOccurred())

			sha256sum, err = util.CalculateSha256ByPersistentCacheTaskID([]*util.PodExec{seedClientPod}, taskID)
			Expect(err).NotTo(HaveOccurred())
			Expect(testFile.GetSha256()).To(Equal(sha256sum))

			sha256sum, err = util.CalculateSha256ByOutput([]*util.PodExec{seedClientPod}, testFile.GetOutputPath())
			Expect(err).NotTo(HaveOccurred())
			Expect(testFile.GetSha256()).To(Equal(sha256sum))
		})
	})

	Context("10MiB file and set export transfer-from-dfdaemon", func() {
		var (
			testFile *util.File
			err      error
		)

		BeforeEach(func() {
			testFile, err = util.GetFileServer().GenerateFile(util.FileSize10MiB)
			Expect(err).NotTo(HaveOccurred())
			Expect(testFile).NotTo(BeNil())
		})

		AfterEach(func() {
			err = util.GetFileServer().DeleteFile(testFile.GetInfo())
			Expect(err).NotTo(HaveOccurred())
		})

		It("import and export should be ok", Label("dfcache", "import and export", "export transfer-from-dfdaemon"), func() {
			var clientPod *util.PodExec
			clientPod, err = util.ClientExec()
			fmt.Println(err)
			Expect(err).NotTo(HaveOccurred())

			_, err := clientPod.Command("sh", "-c", fmt.Sprintf("dfget %s --disable-back-to-source --output %s", testFile.GetDownloadURL(), testFile.GetOutputPath())).CombinedOutput()
			Expect(err).NotTo(HaveOccurred())

			sha256sum, err := util.CalculateSha256ByTaskID([]*util.PodExec{clientPod}, testFile.GetTaskID())
			Expect(err).NotTo(HaveOccurred())
			Expect(testFile.GetSha256()).To(Equal(sha256sum))

			sha256sum, err = util.CalculateSha256ByOutput([]*util.PodExec{clientPod}, testFile.GetOutputPath())
			Expect(err).NotTo(HaveOccurred())
			Expect(testFile.GetSha256()).To(Equal(sha256sum))

			importOut, err := dfcacheImportWithRetry(clientPod, fmt.Sprintf("dfcache import %s --persistent-replica-count 1 --content-for-calculating-task-id %s", testFile.GetOutputPath(), testFile.GetSha256()))
			fmt.Println(string(importOut), err)
			Expect(err).NotTo(HaveOccurred())

			seedClientPod, err := util.SeedClientExec(0)
			fmt.Println(err)
			Expect(err).NotTo(HaveOccurred())

			taskID := idgen.PersistentCacheTaskIDByContent(testFile.GetSha256())
			exportOut, err := seedClientPod.Command("sh", "-c", fmt.Sprintf("dfcache export %s --transfer-from-dfdaemon --output %s", taskID, testFile.GetOutputPath())).CombinedOutput()
			fmt.Println(string(exportOut), err)
			Expect(err).NotTo(HaveOccurred())

			sha256sum, err = util.CalculateSha256ByPersistentCacheTaskID([]*util.PodExec{seedClientPod}, taskID)
			Expect(err).NotTo(HaveOccurred())
			Expect(testFile.GetSha256()).To(Equal(sha256sum))

			sha256sum, err = util.CalculateSha256ByOutput([]*util.PodExec{seedClientPod}, testFile.GetOutputPath())
			Expect(err).NotTo(HaveOccurred())
			Expect(testFile.GetSha256()).To(Equal(sha256sum))
		})
	})
})
