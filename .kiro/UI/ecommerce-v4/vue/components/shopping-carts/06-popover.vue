<template>
  <header class="relative bg-white">
    <nav aria-label="Top" class="mx-auto max-w-7xl sm:px-6 lg:px-8">
      <div class="relative border-b border-gray-200 px-4 pb-14 sm:static sm:px-0 sm:pb-0">
        <div class="flex h-16 items-center justify-between">
          <!-- Logo -->
          <div class="flex flex-1">
            <a href="#">
              <span class="sr-only">Your Company</span>
              <img class="h-8 w-auto" src="https://tailwindcss.com/plus-assets/img/logos/mark.svg?color=indigo&shade=600" alt="" />
            </a>
          </div>

          <div class="absolute inset-x-0 bottom-0 overflow-x-auto border-t border-gray-200 sm:static sm:border-t-0">
            <div class="flex h-14 items-center space-x-8 px-4 sm:h-auto">
              <a v-for="item in navigation" :key="item.name" :href="item.href" class="text-sm font-medium text-gray-700 hover:text-gray-800">{{ item.name }}</a>
            </div>
          </div>

          <div class="flex flex-1 items-center justify-end">
            <!-- Search -->
            <a href="#" class="p-2 text-gray-400 hover:text-gray-500">
              <span class="sr-only">Search</span>
              <MagnifyingGlassIcon class="size-6" aria-hidden="true" />
            </a>

            <!-- Cart -->
            <Popover class="ml-4 flow-root text-sm lg:relative lg:ml-8">
              <PopoverButton class="group -m-2 flex items-center p-2">
                <ShoppingBagIcon class="size-6 shrink-0 text-gray-400 group-hover:text-gray-500" aria-hidden="true" />
                <span class="ml-2 text-sm font-medium text-gray-700 group-hover:text-gray-800">0</span>
                <span class="sr-only">items in cart, view bag</span>
              </PopoverButton>
              <transition enter-active-class="transition ease-out duration-200" enter-from-class="opacity-0" enter-to-class="" leave-active-class="transition ease-in duration-150" leave-from-class="" leave-to-class="opacity-0">
                <PopoverPanel class="absolute top-16 right-0 z-10 mt-px w-full bg-white pb-6 shadow-lg sm:px-2 lg:top-full lg:mt-3 lg:-mr-1.5 lg:w-80 lg:rounded-lg lg:ring-1 lg:ring-black/5">
                  <h2 class="sr-only">Shopping Cart</h2>

                  <form class="mx-auto max-w-2xl px-4">
                    <ul role="list" class="divide-y divide-gray-200">
                      <li v-for="product in products" :key="product.id" class="flex items-center py-6">
                        <img :src="product.imageSrc" :alt="product.imageAlt" class="size-16 flex-none rounded-md border border-gray-200" />
                        <div class="ml-4 flex-auto">
                          <h3 class="font-medium text-gray-900">
                            <a :href="product.href">{{ product.name }}</a>
                          </h3>
                          <p class="text-gray-500">{{ product.color }}</p>
                        </div>
                      </li>
                    </ul>

                    <button type="submit" class="w-full rounded-md border border-transparent bg-indigo-600 px-4 py-2 text-sm font-medium text-white shadow-xs hover:bg-indigo-700 focus:ring-2 focus:ring-indigo-500 focus:ring-offset-2 focus:ring-offset-gray-50 focus:outline-hidden">Checkout</button>

                    <p class="mt-6 text-center">
                      <a href="#" class="text-sm font-medium text-indigo-600 hover:text-indigo-500">View Shopping Bag</a>
                    </p>
                  </form>
                </PopoverPanel>
              </transition>
            </Popover>
          </div>
        </div>
      </div>
    </nav>
  </header>
</template>

<script setup>
import { MagnifyingGlassIcon, ShoppingBagIcon } from '@heroicons/vue/24/outline'
import { Popover, PopoverButton, PopoverPanel } from '@headlessui/vue'

const navigation = [
  { name: 'Women', href: '#' },
  { name: 'Men', href: '#' },
  { name: 'Company', href: '#' },
  { name: 'Stores', href: '#' },
]
const products = [
  {
    id: 1,
    name: 'Throwback Hip Bag',
    href: '#',
    color: 'Salmon',
    imageSrc: 'https://tailwindcss.com/plus-assets/img/ecommerce-images/shopping-cart-page-04-product-01.jpg',
    imageAlt: 'Salmon orange fabric pouch with match zipper, gray zipper pull, and adjustable hip belt.',
  },
  {
    id: 2,
    name: 'Medium Stuff Satchel',
    href: '#',
    color: 'Blue',
    imageSrc: 'https://tailwindcss.com/plus-assets/img/ecommerce-images/shopping-cart-page-04-product-02.jpg',
    imageAlt:
      'Front of satchel with blue canvas body, black straps and handle, drawstring top, and front zipper pouch.',
  },
  {
    id: 3,
    name: 'Zip Tote Basket',
    href: '#',
    color: 'White and black',
    imageSrc: 'https://tailwindcss.com/plus-assets/img/ecommerce-images/shopping-cart-page-04-product-03.jpg',
    imageAlt: 'Front of zip tote bag with white canvas, black canvas straps and handle, and black zipper pulls.',
  },
]
</script>