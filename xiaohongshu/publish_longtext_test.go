package xiaohongshu

import (
	"context"
	"testing"

	"github.com/xpzouying/xiaohongshu-mcp/browser"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPublishLongText(t *testing.T) {

	// t.Skip("SKIP: 测试长文发布")

	b := browser.NewBrowser(false)
	defer b.Close()

	page := b.NewPage()
	defer page.Close()

	action, err := NewPublishLongTextAction(page)
	require.NoError(t, err)

	err = action.PublishLongText(context.Background(), PublishLongTextContent{
		Title:   "测试长文标题",
		Content: "这是一个测试长文的内容，用于验证长文发布功能是否正常工作。包含多行文本内容，模拟真实的长文发布场景。",
	})
	assert.NoError(t, err)
}
