package xiaohongshu

import (
	"context"
	"log/slog"
	"strings"
	"time"
    "os"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
	"github.com/pkg/errors"
)

// PublishImageContent 发布图文内容
type PublishImageContent struct {
	Title      string
	Content    string
	ImagePaths []string
}

type PublishAction struct {
	page *rod.Page
}

const (
	urlOfPublic = `https://creator.xiaohongshu.com/publish/publish?source=official`
)

func NewPublishImageAction(page *rod.Page) (*PublishAction, error) {

	pp := page.Timeout(60 * time.Second)

	pp.MustNavigate(urlOfPublic)

	pp.MustElement(`div.upload-content`).MustWaitVisible()
	slog.Info("wait for upload-content visible success")

	// 等待一段时间确保页面完全加载
	time.Sleep(1 * time.Second)

	createElems := pp.MustElements("div.creator-tab")
	slog.Info("foundcreator-tab elements", "count", len(createElems))
	for _, elem := range createElems {
		text, err := elem.Text()
		if err != nil {
			slog.Error("获取元素文本失败", "error", err)
			continue
		}

		if text == "上传图文" {
			if err := elem.Click(proto.InputMouseButtonLeft, 1); err != nil {
				slog.Error("点击元素失败", "error", err)
				continue
			}
			break
		}
	}

	time.Sleep(1 * time.Second)

	return &PublishAction{
		page: pp,
	}, nil
}

func (p *PublishAction) Publish(ctx context.Context, content PublishImageContent) error {
	if len(content.ImagePaths) == 0 {
		return errors.New("图片不能为空")
	}

	page := p.page.Context(ctx)

	if err := uploadImages(page, content.ImagePaths); err != nil {
		return errors.Wrap(err, "小红书上传图片失败")
	}

	if err := submitPublish(page, content.Title, content.Content); err != nil {
		return errors.Wrap(err, "小红书发布失败")
	}

	return nil
}

func uploadImages(page *rod.Page, imagesPaths []string) error {
	pp := page.Timeout(30 * time.Second)

	// 等待上传输入框出现
	uploadInput := pp.MustElement(".upload-input")

	// 上传多个文件
	uploadInput.MustSetFiles(imagesPaths...)

	// 等待上传完成
	time.Sleep(3 * time.Second)

	return nil
}

func submitPublish(page *rod.Page, title, content string) error {

	titleElem := page.MustElement("div.d-input input")
	titleElem.MustInput(title)

	time.Sleep(1 * time.Second)

	if contentElem, ok := getContentElement(page); ok {
		contentElem.MustInput(content)
	} else {
		return errors.New("没有找到内容输入框")
	}

	time.Sleep(1 * time.Second)

	submitButton := page.MustElement("div.submit div.d-button-content")
	submitButton.MustClick()

	time.Sleep(3 * time.Second)

	return nil
}

// 查找内容输入框 - 使用Race方法处理两种样式
func getContentElement(page *rod.Page) (*rod.Element, bool) {
	var foundElement *rod.Element
	var found bool

	page.Race().
		Element("div.ql-editor").MustHandle(func(e *rod.Element) {
		foundElement = e
		found = true
	}).
		ElementFunc(func(page *rod.Page) (*rod.Element, error) {
			return findTextboxByPlaceholder(page)
		}).MustHandle(func(e *rod.Element) {
		foundElement = e
		found = true
	}).
		MustDo()

	if found {
		return foundElement, true
	}

	slog.Warn("no content element found by any method")
	return nil, false
}

func findTextboxByPlaceholder(page *rod.Page) (*rod.Element, error) {
	elements := page.MustElements("p")
	if elements == nil {
		return nil, errors.New("no p elements found")
	}

	// 查找包含指定placeholder的元素
	placeholderElem := findPlaceholderElement(elements, "输入正文描述")
	if placeholderElem == nil {
		return nil, errors.New("no placeholder element found")
	}

	// 向上查找textbox父元素
	textboxElem := findTextboxParent(placeholderElem)
	if textboxElem == nil {
		return nil, errors.New("no textbox parent found")
	}

	return textboxElem, nil
}

func findPlaceholderElement(elements []*rod.Element, searchText string) *rod.Element {
	for _, elem := range elements {
		placeholder, err := elem.Attribute("data-placeholder")
		if err != nil || placeholder == nil {
			continue
		}

		if strings.Contains(*placeholder, searchText) {
			return elem
		}
	}
	return nil
}

func findTextboxParent(elem *rod.Element) *rod.Element {
	currentElem := elem
	for i := 0; i < 5; i++ {
		parent, err := currentElem.Parent()
		if err != nil {
			break
		}

		role, err := parent.Attribute("role")
		if err != nil || role == nil {
			currentElem = parent
			continue
		}

		if *role == "textbox" {
			return parent
		}

		currentElem = parent
	}
	return nil
}

// NewPublishLongTextAction 创建长文发布Action
func NewPublishLongTextAction(page *rod.Page) (*PublishAction, error) {
	pp := page.Timeout(60 * time.Second)

	pp.MustNavigate(urlOfPublic)
	pp.MustElement(`div.upload-content`).MustWaitVisible()

	// 等待页面加载
	time.Sleep(1 * time.Second)

	// 点击"写长文"选项卡
	createElems := pp.MustElements("div.creator-tab")
	for _, elem := range createElems {
		text, err := elem.Text()
		if err != nil {
			continue
		}
		if text == "写长文" {
			if err := elem.Click(proto.InputMouseButtonLeft, 1); err != nil {
				continue
			}
			break
		}
	}

	time.Sleep(2 * time.Second)

	// 点击"新的创作"按钮
	buttons := pp.MustElements("button")
	var createButton *rod.Element
	for _, btn := range buttons {
		text, err := btn.Text()
		if err != nil {
			continue
		}
		if strings.Contains(text, "新的创作") {
			createButton = btn
			break
		}
	}

	if createButton == nil {
		return nil, errors.New("找不到新的创作按钮")
	}

	if err := createButton.Click(proto.InputMouseButtonLeft, 1); err != nil {
		return nil, errors.Wrap(err, "点击新的创作按钮失败")
	}

	// 等待页面跳转
	time.Sleep(2 * time.Second)

	return &PublishAction{
		page: pp,
	}, nil
}

// PublishLongText 发布长文
func (p *PublishAction) PublishLongText(ctx context.Context, content PublishLongTextContent) error {
	if content.Title == "" || content.Content == "" {
		return errors.New("标题和内容不能为空")
	}

	page := p.page.Context(ctx)

	if err := submitLongTextPublish(page, content.Title, content.Content); err != nil {
		return errors.Wrap(err, "小红书长文发布失败")
	}

	return nil
}

// submitLongTextPublish 提交长文发布
func submitLongTextPublish(page *rod.Page, title, content string) error {
	pp := page.Timeout(30 * time.Second)

	// 填写标题
	titleElem, err := findLongTextTitleElement(pp)
	if err != nil {
		return errors.Wrap(err, "找不到标题输入框")
	}

	titleElem.MustClick()
	time.Sleep(500 * time.Millisecond)
	titleElem.MustSelectAllText()
	titleElem.MustInput(title)
	time.Sleep(1 * time.Second)

	// 填写内容
	contentElem, err := findLongTextContentElement(pp)
	if err != nil {
		return errors.Wrap(err, "找不到内容输入区域")
	}

	contentElem.MustClick()
	time.Sleep(500 * time.Millisecond)
	contentElem.MustInput(content)
	time.Sleep(1 * time.Second)

	// 点击"一键排版"按钮
	oneClickFormatButton, err := findOneClickFormatButton(pp)
	if err != nil {
		return errors.Wrap(err, "找不到一键排版按钮")
	}
	oneClickFormatButton.MustClick()
	time.Sleep(2 * time.Second)

	// 点击"下一步"按钮
	nextStepButton, err := findNextStepButton(pp)
	if err != nil {
		return errors.Wrap(err, "找不到下一步按钮")
	}
	nextStepButton.MustClick()
	time.Sleep(3 * time.Second)

	// 等待确认页面加载
	time.Sleep(5 * time.Second)

	// 在确认页面重新填写标题和内容
	if err := fillConfirmationPage(pp, title, content); err != nil {
		return errors.Wrap(err, "填写确认页面失败")
	}

	// 设置可见范围为仅自己可见
	if err := setVisibilityToPrivate(pp); err != nil {
		return errors.Wrap(err, "设置可见范围失败")
	}

	// 点击发布按钮
	publishButton, err := findPublishButton(pp)
	if err != nil {
		return errors.Wrap(err, "找不到发布按钮")
	}
	publishButton.MustClick()
	time.Sleep(3 * time.Second)

	return nil
}

// fillConfirmationPage 在确认页面填写标题和内容
func fillConfirmationPage(page *rod.Page, title, content string) error {
	// 填写确认页面的标题
	confirmTitleElem, err := findConfirmationTitleElement(page)
	if err != nil {
		return errors.Wrap(err, "找不到确认页面标题输入框")
	}

	confirmTitleElem.MustClick()
	time.Sleep(500 * time.Millisecond)
	confirmTitleElem.MustSelectAllText()
	confirmTitleElem.MustInput(title)
	time.Sleep(1 * time.Second)

	// 填写确认页面的内容
	confirmContentElem, err := findConfirmationContentElement(page)
	if err != nil {
		return errors.Wrap(err, "找不到确认页面内容输入区域")
	}

	confirmContentElem.MustClick()
	time.Sleep(500 * time.Millisecond)
	// ProseMirror编辑器不支持MustSelectAllText，直接输入内容
	confirmContentElem.MustInput(content)
	time.Sleep(1 * time.Second)

	return nil
}

// setVisibilityToPrivate 设置可见范围为仅自己可见
func setVisibilityToPrivate(page *rod.Page) error {
	// 等待页面完全加载
	time.Sleep(1 * time.Second)

	// 滚动到页面底部，确保设置区域可见
	time.Sleep(1 * time.Second)


	// 查找可见范围选择器
	visibilitySelector, err := findVisibilitySelector(page)
	if err != nil {
		return errors.Wrap(err, "找不到可见范围选择器")
	}

	// 点击选择器展开下拉菜单
	visibilitySelector.MustScrollIntoView()
	visibilitySelector.MustClick()
	time.Sleep(500 * time.Millisecond) // 等待下拉菜单展开

	// 查找并点击"仅自己可见"选项
	privateOption, err := findPrivateVisibilityOption(page)
	if err != nil {
		return errors.Wrap(err, "找不到仅自己可见选项")
	}

	privateOption.MustClick()
	return nil
}

// findLongTextTitleElement 查找长文标题输入框
func findLongTextTitleElement(page *rod.Page) (*rod.Element, error) {
	time.Sleep(1 * time.Second)

	// 查找包含"输入标题"文本的元素
	titleElements := page.MustElements("div, span, input, textarea")
	for _, elem := range titleElements {
		text, _ := elem.Text()
		if strings.Contains(text, "输入标题") {
			// 检查这个元素本身是否可编辑
			contentEditable, _ := elem.Attribute("contenteditable")
			if contentEditable != nil && *contentEditable == "true" {
				return elem, nil
			}

			// 查找父元素中的可编辑元素
			parent, err := elem.Parent()
			if err == nil {
				editableChildren := parent.MustElements("[contenteditable='true']")
				if len(editableChildren) > 0 {
					return editableChildren[0], nil
				}
			}
		}
	}

	// 降级策略：使用第一个可编辑元素作为标题输入框
	editableDivs := page.MustElements("div[contenteditable='true'], input[type='text'], textarea")
	if len(editableDivs) > 0 {
		return editableDivs[0], nil
	}

	return nil, errors.New("找不到标题输入框")
}

// findLongTextContentElement 查找长文内容输入区域
func findLongTextContentElement(page *rod.Page) (*rod.Element, error) {
	time.Sleep(1 * time.Second)

	// 查找TipTap富文本编辑器
	editableDivs := page.MustElements("div[contenteditable='true']")
	for _, div := range editableDivs {
		className, _ := div.Attribute("class")
		if className != nil {
			classStr := *className
			if strings.Contains(classStr, "ProseMirror") || strings.Contains(classStr, "tiptap") {
				return div, nil
			}
		}
	}

	// 降级策略：使用第一个可编辑元素
	if len(editableDivs) > 0 {
		return editableDivs[0], nil
	}

	return nil, errors.New("找不到内容输入区域")
}

// findOneClickFormatButton 查找一键排版按钮
func findOneClickFormatButton(page *rod.Page) (*rod.Element, error) {
	buttons := page.MustElements("button")
	for _, btn := range buttons {
		text, err := btn.Text()
		if err != nil {
			continue
		}
		if strings.Contains(text, "一键排版") {
			return btn, nil
		}
	}
	return nil, errors.New("找不到一键排版按钮")
}

// findNextStepButton 查找下一步按钮
func findNextStepButton(page *rod.Page) (*rod.Element, error) {
	buttons := page.MustElements("button")
	for _, btn := range buttons {
		text, err := btn.Text()
		if err != nil {
			continue
		}
		if strings.Contains(text, "下一步") {
			return btn, nil
		}
	}
	return nil, errors.New("找不到下一步按钮")
}

// findPublishButton 查找发布按钮
func findPublishButton(page *rod.Page) (*rod.Element, error) {
	buttons := page.MustElements("button")
	for _, btn := range buttons {
		text, err := btn.Text()
		if err != nil {
			continue
		}
		if text == "发布" || strings.Contains(text, "发布") {
			return btn, nil
		}
	}
	return nil, errors.New("找不到发布按钮")
}

// findConfirmationTitleElement 查找确认页面的标题输入框
func findConfirmationTitleElement(page *rod.Page) (*rod.Element, error) {
	time.Sleep(1 * time.Second)

	// 查找所有输入框元素
	allElements := page.MustElements("input, textarea, [contenteditable='true']")
	for _, elem := range allElements {
		// 检查是否是标题相关的输入框
		placeholder, _ := elem.Attribute("placeholder")
		if placeholder != nil && strings.Contains(*placeholder, "标题") {
			return elem, nil
		}

		// 检查aria-label
		ariaLabel, _ := elem.Attribute("aria-label")
		if ariaLabel != nil && strings.Contains(*ariaLabel, "标题") {
			return elem, nil
		}

		// 检查父元素的文本内容
		parent, err := elem.Parent()
		if err == nil {
			parentText, _ := parent.Text()
			if strings.Contains(parentText, "标题") || strings.Contains(parentText, "输入标题") {
				return elem, nil
			}
		}
	}

	// 降级策略：使用第一个输入框作为标题框
	if len(allElements) > 0 {
		return allElements[0], nil
	}

	return nil, errors.New("找不到确认页面标题输入框")
}

// findConfirmationContentElement 查找确认页面的内容输入区域
func findConfirmationContentElement(page *rod.Page) (*rod.Element, error) {
	time.Sleep(1 * time.Second)

	// 首先查找富文本编辑器
	editableDivs := page.MustElements("div[contenteditable='true']")
	for _, div := range editableDivs {
		className, _ := div.Attribute("class")
		if className != nil {
			classStr := *className
			// 查找ProseMirror或tiptap编辑器
			if strings.Contains(classStr, "ProseMirror") || strings.Contains(classStr, "tiptap") {
				return div, nil
			}
		}

		// 检查是否包含内容相关的属性
		ariaLabel, _ := div.Attribute("aria-label")
		if ariaLabel != nil && (strings.Contains(*ariaLabel, "内容") || strings.Contains(*ariaLabel, "正文")) {
			return div, nil
		}
	}

	// 查找textarea元素
	textareas := page.MustElements("textarea")
	for _, textarea := range textareas {
		placeholder, _ := textarea.Attribute("placeholder")
		if placeholder != nil && (strings.Contains(*placeholder, "内容") || strings.Contains(*placeholder, "正文")) {
			return textarea, nil
		}
	}

	// 降级策略：使用最后一个可编辑元素作为内容区域（通常是最大的）
	if len(editableDivs) > 0 {
		return editableDivs[len(editableDivs)-1], nil
	}

	return nil, errors.New("找不到确认页面内容输入区域")
}

func findVisibilitySelector(page *rod.Page) (*rod.Element, error) {
    // 滚动到底部确保设置区域可见
    page.Mouse.Scroll(0.0, 10000.0, 1) // 正确的参数类型
    page.MustWaitIdle()

    el, err := page.Timeout(5 * time.Second).ElementR("div,span", "可见范围|公开可见|谁可以看")
    if err != nil || el == nil {
        slog.Error("找不到可见范围选择器", "err", err)
        return nil, errors.New("找不到可见范围选择器")
    }

    el.MustScrollIntoView()
    txt, _ := el.Text()
    slog.Info("找到的可见范围元素内容", "text", txt)
    return el, nil
}

// findPrivateVisibilityOption 查找"仅自己可见"选项
func findPrivateVisibilityOption(page *rod.Page) (*rod.Element, error) {
	// 基于DOM结构：下拉菜单展开后，查找包含"仅自己可见"的选项
	// 等待下拉菜单完全展开
	time.Sleep(500 * time.Millisecond)


	// 精确匹配"仅自己可见"文本
	textElements := page.MustElements("div, span")
	for _, elem := range textElements {
		text, _ := elem.Text()
		// 精确匹配，根据DOM结构应该是generic [ref=e742]: "仅自己可见"
		if text == "仅自己可见" {
			return elem, nil
		}
	}

	// 备选策略：包含匹配
	for _, elem := range textElements {
		text, _ := elem.Text()
		if strings.Contains(text, "仅自己可见") {
			return elem, nil
		}
	}

	return nil, errors.New("找不到仅自己可见选项")
}

