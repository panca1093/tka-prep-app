/**
 * @vitest-environment jsdom
 */
import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import RichTextViewer from '../editor/RichTextViewer.vue'

describe('RichTextViewer', () => {
  it('renders plain text unchanged', () => {
    const wrapper = mount(RichTextViewer, { props: { html: 'Hello world' } })
    expect(wrapper.text()).toContain('Hello world')
  })

  it('renders bold text', () => {
    const wrapper = mount(RichTextViewer, { props: { html: '<b>bold</b>' } })
    expect(wrapper.html()).toContain('<b>bold</b>')
  })

  it('strips script tags', () => {
    const wrapper = mount(RichTextViewer, { props: { html: '<script>alert(1)</script><p>safe</p>' } })
    expect(wrapper.html()).not.toContain('<script>')
    expect(wrapper.html()).not.toContain('alert(1)')
    expect(wrapper.text()).toContain('safe')
  })

  it('strips onclick handlers', () => {
    const wrapper = mount(RichTextViewer, { props: { html: '<p onclick="alert(1)">text</p>' } })
    expect(wrapper.html()).not.toContain('onclick')
    expect(wrapper.text()).toContain('text')
  })

  it('allows img with /uploads/ src', () => {
    const wrapper = mount(RichTextViewer, { props: { html: '<img src="/uploads/abc123.png" alt="diagram">' } })
    expect(wrapper.html()).toContain('/uploads/abc123.png')
  })

  it('allows img with https src', () => {
    const wrapper = mount(RichTextViewer, { props: { html: '<img src="https://example.com/diagram.jpg">' } })
    expect(wrapper.html()).toContain('example.com')
  })

  it('preserves data-formula spans', () => {
    const wrapper = mount(RichTextViewer, { props: { html: '<span data-formula="F(x)=x-1"></span>' } })
    expect(wrapper.html()).toContain('data-formula="F(x)=x-1"')
  })

  it('renders empty string without error', () => {
    const wrapper = mount(RichTextViewer, { props: { html: '' } })
    expect(wrapper.find('.richtext-viewer').exists()).toBe(true)
    expect(wrapper.text()).toBe('')
  })
})
