package terraform

import "fmt"

func (c *Client) CreateALL(plan bool) (string, error) {
	if plan {
		return c.plan("")
	}
	return c.apply("")
}

func (c *Client) CreateTeam(plan bool) (string, error) {
	if plan {
		return c.plan("")
	}
	return c.apply("")
}

func (c *Client) CreateFromProblemId(problemId string, plan bool) (string, error) {
	str := fmt.Sprintf("-target=module.%s", problemId)
	if plan {
		return c.plan(str)
	}
	return c.apply(str)
}

func (c *Client) DestroyTeam(plan bool) (string, error) {
	if plan {
		return c.plan("-destroy")
	}
	return c.apply("-destroy")
}

func (c *Client) DestroyFromProblemId(problemId string, plan bool) (string, error) {
	str := fmt.Sprintf("-destroy -target=module.%s", problemId)
	if plan {
		return c.plan(str)
	}
	return c.apply(str)
}

func (c *Client) RecreateTeam(plan bool) (string, int, error) {
	_, resourceCount, err := c.GetResourceTargetId("")
	if err != nil {
		return "", resourceCount, err
	}

	if _, err := c.DestroyTeam(plan); err != nil {
		return "", resourceCount, err
	}
	result, err := c.CreateTeam(plan)
	return result, resourceCount, err
}

func (c *Client) RecreateFromProblemId(problemId string, plan bool) (string, int, error) {
	str, resourceCount, err := c.GetResourceTargetId("module." + problemId)
	/*
		apply --target resource idがない時Applyを試行
	*/
	if err != nil {
		targetOption := fmt.Sprintf(" --target module.%s ", problemId)
		result, err := c.apply(targetOption)
		return result, 0, err
	}
	str += fmt.Sprintf(" --target module.%s ", problemId)
	if plan {
		result, err := c.plan(str)
		return result, resourceCount, err
	}
	result, err := c.apply(str)

	return result, resourceCount, err
}

func (c *Client) RecreateFromRaw(replaceOpt string, plan bool) (string, error) {
	if plan {
		result, err := c.plan(replaceOpt)
		return result, err
	}
	result, err := c.apply(replaceOpt)
	return result, err
}
