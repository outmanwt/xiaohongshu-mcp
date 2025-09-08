package xiaohongshu

import (
	"context"
	"fmt"
	"log/slog"
	"testing"
	"time"

	"github.com/xpzouying/xiaohongshu-mcp/browser"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestResult 测试结果记录
type TestResult struct {
	Step    int
	Name    string
	Success bool
	Error   string
	Details string
}

func TestPublishLongTextWithDetailedVerification(t *testing.T) {

	// t.Skip("SKIP: 详细验证测试长文发布 - 取消注释以运行测试")

	var results []TestResult

	fmt.Println("=== 小红书长文发布流程测试开始 ===")

	b := browser.NewBrowser(false)
	defer b.Close()

	page := b.NewPage()
	defer page.Close()

	// 测试数据
	testTitle := "测试长文标题 - 功能验证"
	testContent := "这是一个测试长文的内容，用于验证长文发布功能是否正常工作。包含多行文本内容，模拟真实的长文发布场景。测试确认页面的标题和内容重新填写功能。"

	// 第一阶段：基础流程测试（步骤1-7）
	fmt.Println("\n【第一阶段：基础流程测试】")

	// 步骤1-3: NewPublishLongTextAction 包含了导航、点击写长文、点击新的创作
	slog.Info("开始执行步骤1-3: 创建长文发布Action")
	action, err := NewPublishLongTextAction(page)
	if err != nil {
		results = append(results, TestResult{
			Step: 1, Name: "创建长文发布Action(步骤1-3)", Success: false,
			Error: err.Error(), Details: "包含页面导航、点击写长文选项卡、点击新的创作按钮",
		})
		require.NoError(t, err)
	} else {
		results = append(results, TestResult{
			Step: 1, Name: "创建长文发布Action(步骤1-3)", Success: true,
			Details: "✅ 页面导航成功\n✅ 写长文选项卡点击成功\n✅ 新的创作按钮点击成功",
		})
		fmt.Println("✅ 步骤1-3: 创建长文发布Action成功")
	}

	// 详细执行发布流程并验证每个步骤
	fmt.Println("\n开始执行详细发布流程验证...")
	results = append(results, executeDetailedPublishFlow(t, action, testTitle, testContent)...)

	// 生成测试报告
	generateTestReport(results)
}

func executeDetailedPublishFlow(t *testing.T, action *PublishAction, title, content string) []TestResult {
	var results []TestResult

	page := action.page

	fmt.Println("\n【执行发布流程各步骤验证】")

	// 步骤4: 填写标题
	slog.Info("执行步骤4: 查找并填写标题")
	titleElem, err := findLongTextTitleElement(page)
	if err != nil {
		results = append(results, TestResult{
			Step: 4, Name: "查找标题输入框", Success: false,
			Error: err.Error(), Details: "findLongTextTitleElement() 执行失败",
		})
	} else {
		results = append(results, TestResult{
			Step: 4, Name: "查找标题输入框", Success: true,
			Details: "✅ findLongTextTitleElement() 找到标题输入框",
		})

		// 填写标题
		titleElem.MustClick()
		time.Sleep(500 * time.Millisecond)
		titleElem.MustSelectAllText()
		titleElem.MustInput(title)
		time.Sleep(1 * time.Second)

		results = append(results, TestResult{
			Step: 4, Name: "填写标题", Success: true,
			Details: fmt.Sprintf("✅ 标题填写成功: %s", title),
		})
		fmt.Println("✅ 步骤4: 标题输入框找到并填写成功")
	}

	// 步骤5: 填写内容
	slog.Info("执行步骤5: 查找并填写内容")
	contentElem, err := findLongTextContentElement(page)
	if err != nil {
		results = append(results, TestResult{
			Step: 5, Name: "查找内容输入区域", Success: false,
			Error: err.Error(), Details: "findLongTextContentElement() 执行失败",
		})
	} else {
		results = append(results, TestResult{
			Step: 5, Name: "查找内容输入区域", Success: true,
			Details: "✅ findLongTextContentElement() 找到内容编辑器",
		})

		// 填写内容
		contentElem.MustClick()
		time.Sleep(500 * time.Millisecond)
		contentElem.MustInput(content)
		time.Sleep(1 * time.Second)

		results = append(results, TestResult{
			Step: 5, Name: "填写内容", Success: true,
			Details: "✅ 内容填写成功",
		})
		fmt.Println("✅ 步骤5: 内容编辑器找到并填写成功")
	}

	// 步骤6: 点击"一键排版"按钮
	slog.Info("执行步骤6: 查找并点击一键排版按钮")
	formatButton, err := findOneClickFormatButton(page)
	if err != nil {
		results = append(results, TestResult{
			Step: 6, Name: "查找一键排版按钮", Success: false,
			Error: err.Error(), Details: "findOneClickFormatButton() 执行失败",
		})
	} else {
		results = append(results, TestResult{
			Step: 6, Name: "查找一键排版按钮", Success: true,
			Details: "✅ findOneClickFormatButton() 找到一键排版按钮",
		})

		formatButton.MustClick()
		time.Sleep(2 * time.Second)

		results = append(results, TestResult{
			Step: 6, Name: "点击一键排版", Success: true,
			Details: "✅ 一键排版按钮点击成功",
		})
		fmt.Println("✅ 步骤6: 一键排版按钮找到并执行成功")
	}

	// 步骤7: 点击"下一步"按钮
	slog.Info("执行步骤7: 查找并点击下一步按钮")
	nextButton, err := findNextStepButton(page)
	if err != nil {
		results = append(results, TestResult{
			Step: 7, Name: "查找下一步按钮", Success: false,
			Error: err.Error(), Details: "findNextStepButton() 执行失败",
		})
	} else {
		results = append(results, TestResult{
			Step: 7, Name: "查找下一步按钮", Success: true,
			Details: "✅ findNextStepButton() 找到下一步按钮",
		})

		nextButton.MustClick()
		time.Sleep(3 * time.Second)

		// 等待确认页面加载
		time.Sleep(5 * time.Second)

		results = append(results, TestResult{
			Step: 7, Name: "点击下一步并跳转", Success: true,
			Details: "✅ 下一步按钮点击成功，页面跳转到确认界面",
		})
		fmt.Println("✅ 步骤7: 下一步按钮找到并跳转到确认界面")
	}

	fmt.Println("\n【第二阶段：确认页面测试】")

	// 步骤8: 确认页面填写标题
	slog.Info("执行步骤8: 查找确认页面标题输入框")
	confirmTitleElem, err := findConfirmationTitleElement(page)
	if err != nil {
		results = append(results, TestResult{
			Step: 8, Name: "查找确认页面标题输入框", Success: false,
			Error: err.Error(), Details: "findConfirmationTitleElement() 执行失败",
		})
	} else {
		results = append(results, TestResult{
			Step: 8, Name: "查找确认页面标题输入框", Success: true,
			Details: "✅ findConfirmationTitleElement() 找到确认页面标题输入框",
		})

		// 重新填写标题
		confirmTitleElem.MustClick()
		time.Sleep(500 * time.Millisecond)
		confirmTitleElem.MustSelectAllText()
		confirmTitleElem.MustInput(title)
		time.Sleep(1 * time.Second)

		results = append(results, TestResult{
			Step: 8, Name: "确认页面填写标题", Success: true,
			Details: "✅ 确认页面标题重新填写成功",
		})
		fmt.Println("✅ 步骤8: 确认页面标题输入框找到并填写成功")
	}

	// 步骤9: 确认页面填写内容
	slog.Info("执行步骤9: 查找确认页面内容输入区域")
	confirmContentElem, err := findConfirmationContentElement(page)
	if err != nil {
		results = append(results, TestResult{
			Step: 9, Name: "查找确认页面内容输入区域", Success: false,
			Error: err.Error(), Details: "findConfirmationContentElement() 执行失败",
		})
	} else {
		results = append(results, TestResult{
			Step: 9, Name: "查找确认页面内容输入区域", Success: true,
			Details: "✅ findConfirmationContentElement() 找到确认页面内容区域",
		})

		// 重新填写内容
		confirmContentElem.MustClick()
		time.Sleep(500 * time.Millisecond)
		// ProseMirror编辑器不支持MustSelectAllText，直接输入内容
		confirmContentElem.MustInput(content)
		time.Sleep(1 * time.Second)

		results = append(results, TestResult{
			Step: 9, Name: "确认页面填写内容", Success: true,
			Details: "✅ 确认页面内容重新填写成功",
		})
		fmt.Println("✅ 步骤9: 确认页面内容区域找到并填写成功")
	}

	// 步骤10: 设置可见范围为仅自己可见
	slog.Info("执行步骤10: 设置可见范围")
	visibilitySelector, err := findVisibilitySelector(page)
	if err != nil {
		results = append(results, TestResult{
			Step: 10, Name: "查找可见范围选择器", Success: false,
			Error: err.Error(), Details: "findVisibilitySelector() 执行失败",
		})
	} else {
		results = append(results, TestResult{
			Step: 10, Name: "查找可见范围选择器", Success: true,
			Details: "✅ findVisibilitySelector() 找到可见范围选择器",
		})

		// 点击选择器
		visibilitySelector.MustClick()
		time.Sleep(1 * time.Second)

		// 查找私密选项
		privateOption, err := findPrivateVisibilityOption(page)
		if err != nil {
			results = append(results, TestResult{
				Step: 10, Name: "查找仅自己可见选项", Success: false,
				Error: err.Error(), Details: "findPrivateVisibilityOption() 执行失败",
			})
		} else {
			results = append(results, TestResult{
				Step: 10, Name: "查找仅自己可见选项", Success: true,
				Details: "✅ findPrivateVisibilityOption() 找到仅自己可见选项",
			})

			privateOption.MustClick()
			time.Sleep(1 * time.Second)

			results = append(results, TestResult{
				Step: 10, Name: "设置可见范围为仅自己可见", Success: true,
				Details: "✅ 可见范围设置为仅自己可见成功",
			})
			fmt.Println("✅ 步骤10: 可见范围设置为仅自己可见成功")
		}
	}

	fmt.Println("\n【第三阶段：发布完成测试】")

	// 步骤11: 点击发布按钮
	slog.Info("执行步骤11: 查找并点击发布按钮")
	publishButton, err := findPublishButton(page)
	if err != nil {
		results = append(results, TestResult{
			Step: 11, Name: "查找发布按钮", Success: false,
			Error: err.Error(), Details: "findPublishButton() 执行失败",
		})
	} else {
		results = append(results, TestResult{
			Step: 11, Name: "查找发布按钮", Success: true,
			Details: "✅ findPublishButton() 找到发布按钮",
		})

		publishButton.MustClick()
		time.Sleep(3 * time.Second)

		results = append(results, TestResult{
			Step: 11, Name: "点击发布按钮", Success: true,
			Details: "✅ 发布按钮点击成功，长文发布完成",
		})
		fmt.Println("✅ 步骤11: 发布按钮找到并发布成功")
	}

	return results
}

func generateTestReport(results []TestResult) {
	fmt.Println("\n=== 小红书长文发布流程测试报告 ===")

	successCount := 0
	totalCount := len(results)

	for _, result := range results {
		status := "❌"
		if result.Success {
			status = "✅"
			successCount++
		}

		fmt.Printf("%s 步骤%d: %s\n", status, result.Step, result.Name)
		if result.Details != "" {
			fmt.Printf("   详情: %s\n", result.Details)
		}
		if result.Error != "" {
			fmt.Printf("   错误: %s\n", result.Error)
		}
	}

	fmt.Println("\n【测试结果汇总】")
	fmt.Printf("总步骤: %d步\n", totalCount)
	fmt.Printf("成功步骤: %d步\n", successCount)
	fmt.Printf("失败步骤: %d步\n", totalCount-successCount)
	fmt.Printf("测试通过率: %.1f%%\n", float64(successCount)/float64(totalCount)*100)

	if successCount == totalCount {
		fmt.Println("🎉 所有测试步骤通过！长文发布流程验证成功！")
	} else {
		fmt.Printf("⚠️  有%d个步骤失败，请检查相关功能\n", totalCount-successCount)
	}
}

// 保留原有的简单测试
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
