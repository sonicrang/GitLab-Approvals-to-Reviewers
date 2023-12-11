package main

import (
	"fmt"
	"reflect"
	"sort"

	"github.com/xanzy/go-gitlab"
)

func InitAPI() *gitlab.Client {
	git, err := gitlab.NewClient(GITLAB_ACCESS_TOKEN, gitlab.WithBaseURL(GITLAB_URL+"/api/v4"))
	if err != nil {
		logger.Error("failed to create client: " + err.Error())
		return nil
	}
	return git
}

func ApprovalstoReviewer(git *gitlab.Client, pid int, mriid int) {

	// get reviewers in mr
	mr, _, err := git.MergeRequests.GetMergeRequest(pid, mriid, &gitlab.GetMergeRequestsOptions{})
	if err != nil {
		logger.Error(fmt.Sprintf("cloud not get reviewers in pid %v mriid %v", pid, mriid) + err.Error())
	}
	reviewer_ids := []int{}
	for _, user := range mr.Reviewers {
		reviewer_ids = append(reviewer_ids, user.ID)
	}
	sort.Ints(reviewer_ids)
	logger.Info(fmt.Sprintf("reviewers in pid %v mriid %v: %+v", pid, mriid, reviewer_ids))

	// get approvals in mr
	approvals, _, err := git.MergeRequests.GetMergeRequestApprovals(pid, mriid)
	if err != nil {
		logger.Error(fmt.Sprintf("cloud not get approvals in pid %v mriid %v", pid, mriid) + err.Error())
	}

	approval_ids := []int{}
	// get approved users in approvals
	for _, user := range approvals.ApprovedBy {
		approval_ids = append(approval_ids, user.User.ID)
	}
	// get unapproved users in approvals
	for _, user := range approvals.SuggestedApprovers {
		approval_ids = append(approval_ids, user.ID)
	}
	sort.Ints(approval_ids)
	logger.Info(fmt.Sprintf("approvals in pid %v mriid %v: %+v", pid, mriid, approval_ids))

	// if reviewers not equals approvals, use approvals to replace reviewers
	if !reflect.DeepEqual(reviewer_ids, approval_ids) {
		_, _, err = git.MergeRequests.UpdateMergeRequest(pid, mriid, &gitlab.UpdateMergeRequestOptions{ReviewerIDs: &approval_ids})
		if err != nil {
			logger.Error(fmt.Sprintf("cloud not update merge_request in pid %v mriid %v", pid, mriid) + err.Error())
		} else {
			logger.Info(fmt.Sprintf("use approvals to replace reviewers in pid %v mriid %v", pid, mriid))
		}
	}

}
